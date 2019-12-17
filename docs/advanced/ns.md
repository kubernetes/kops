### Getting the value of your subdomain NS records in AWS

 - Note your hosted zone ID

```bash
HZC=$(aws route53 list-hosted-zones | jq -r '.HostedZones[] | select(.Name=="subdomain.example.com.") | .Id' | tee /dev/stderr)

```

  - Note your nameservers for the subdomain

```bash
aws route53 get-hosted-zone --id $HZC | jq .DelegationSet.NameServers
```
