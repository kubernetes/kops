# kops & MFA

You can secure `kops` with MFA by creating an AWS role & policy that requires MFA to access to the `KOPS_STATE_STORE` bucket. Unfortunately the Go AWS SDK does not transparently support assuming roles with required MFA. This may change in a future version. `kops` plans to support this behavior eventually. You can track progress in this [Github issue](https://github.com/kubernetes/kops/issues/226). If you'd like to use MFA with `kops`, you'll need a work around until then.

## The Workaround #1

The work around uses `aws sts assume-role` in combination with an MFA prompt to retrieve temporary AWS access keys. This provides `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN` environment variables which are automatically picked up by Go AWS SDK. You provide the MFA & Role ARNs, then invoke `kops`.

Here's an example wrapper script:

```bash
#!/usr/bin/env bash

set -euo pipefail

main() {
	local role_arn="${KOPS_MFA_ROLE_ARN:-}"
	local serial_number="${KOPS_MFA_ARN:-}"
	local token_code

	if [ -z "${role_arn}" ]; then
		echo "Set the KOPS_MFA_ROLE_ARN environment variable" 1>&2
		return 1
	fi

	if [ -z "${serial_number}" ]; then
		echo "Set the KOPS_MFA_ARN environment variable" 1>&2
		return 1
	fi

	echo -n "Enter MFA Code: "
	read -s token_code

	# NOTE: The keys should not be exported as AWS_ACCESS_KEY_ID
	# or AWS_SECRET_ACCESS_KEY_ID. This will not work. They
	# should be exported as other names which can be used below. This prevents
	# them from incorrectly being picked up from libraries or commands.
	temporary_credentials="$(aws \
		sts assume-role \
		--role-arn="${role_arn}" \
		--serial-number="${serial_number}" \
		--token-code="${token_code}" \
		--role-session-name="kops-access"
	)"

	unset AWS_PROFILE

	export "AWS_ACCESS_KEY_ID=$(echo "${temporary_credentials}" | jq -re '.Credentials.AccessKeyId')"
	export "AWS_SECRET_ACCESS_KEY=$(echo "${temporary_credentials}" | jq -re '.Credentials.SecretAccessKey')"
	export "AWS_SESSION_TOKEN=$(echo "${temporary_credentials}" | jq -re '.Credentials.SessionToken')"

	exec kops "$@"
}

main "$@"
```

#### Usage

Download the script as `kops-mfa`, make it executable, put it on `$PATH`, set the `KOPS_MFA_ARN` and `KOPS_MFA_ROLE_ARN` environment variables. Run as `kops-mfa` followed by any `kops` command.


## The Workaround #2
Use [awsudo](https://github.com/makethunder/awsudo) to generate temp credentials. This is similar to previous but shorter:
```
pip install awsudo
env $(awsudo ${AWS_PROFILE} | grep AWS | xargs) kops ...
```
