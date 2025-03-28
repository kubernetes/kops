# v1.58.0 (2025-03-24)

* **Feature**: This release adds the AvailableSecurityUpdatesComplianceStatus field to patch baseline operations, as well as the AvailableSecurityUpdateCount and InstancesWithAvailableSecurityUpdates to patch state operations. Applies to Windows Server managed nodes only.

# v1.57.2 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.57.1 (2025-02-28)

* **Documentation**: Systems Manager doc-only updates for Feb. 2025.

# v1.57.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.13 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.12 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.11 (2025-02-04)

* No change notes available for this release.

# v1.56.10 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.9 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.8 (2025-01-24)

* **Documentation**: Systems Manager doc-only update for January, 2025.
* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.56.7 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.56.6 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.5 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.56.4 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.3 (2025-01-08)

* No change notes available for this release.

# v1.56.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.56.0 (2024-11-21)

* **Feature**: Added support for providing high-level overviews of managed nodes and previewing the potential impact of a runbook execution.

# v1.55.6 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.5 (2024-11-07)

* **Bug Fix**: Adds case-insensitive handling of error message fields in service responses

# v1.55.4 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.3 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.55.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.54.4 (2024-10-03)

* No change notes available for this release.

# v1.54.3 (2024-09-27)

* No change notes available for this release.

# v1.54.2 (2024-09-25)

* No change notes available for this release.

# v1.54.1 (2024-09-23)

* No change notes available for this release.

# v1.54.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.53.0 (2024-09-17)

* **Feature**: Support for additional levels of cross-account, cross-Region organizational units in Automation. Various documentation updates.
* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.52.8 (2024-09-04)

* No change notes available for this release.

# v1.52.7 (2024-09-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.6 (2024-08-22)

* No change notes available for this release.

# v1.52.5 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.4 (2024-08-09)

* **Documentation**: Systems Manager doc-only updates for August 2024.

# v1.52.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.52.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.51.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.51.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.7 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.6 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.5 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.4 (2024-05-23)

* No change notes available for this release.

# v1.50.3 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.2 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.50.1 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.50.0 (2024-04-24)

* **Feature**: Add SSM DescribeInstanceProperties API to public AWS SDK.

# v1.49.5 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.4 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.3 (2024-03-12)

* **Documentation**: March 2024 doc-only updates for Systems Manager.

# v1.49.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.49.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.48.0 (2024-02-21)

* **Feature**: This release adds support for sharing Systems Manager parameters with other AWS accounts.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.47.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.47.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.46.1 (2024-02-15)

* **Bug Fix**: Correct failure to determine the error type in awsJson services that could occur when errors were modeled with a non-string `code` field.

# v1.46.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.45.0 (2024-01-31)

* **Feature**: This release adds an optional Duration parameter to StateManager Associations. This allows customers to specify how long an apply-only-on-cron association execution should run. Once the specified Duration is out all the ongoing cancellable commands or automations are cancelled.

# v1.44.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.6 (2023-12-20)

* No change notes available for this release.

# v1.44.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.44.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.44.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.44.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.3 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.2 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.43.1 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2023-11-16)

* **Feature**: This release introduces the ability to filter automation execution steps which have parent steps. In addition, runbook variable information is returned by GetAutomationExecution and parent step information is returned by the DescribeAutomationStepExecutions API.

# v1.42.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.42.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.39.0 (2023-10-20)

* **Feature**: This release introduces a new API: DeleteOpsItem. This allows deletion of an OpsItem.

# v1.38.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.0 (2023-09-25)

* **Feature**: This release updates the enum values for ResourceType in SSM DescribeInstanceInformation input and ConnectionStatus in GetConnectionStatus output.

# v1.37.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2023-08-01)

* No change notes available for this release.

# v1.37.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.9 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.8 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.7 (2023-06-27)

* **Documentation**: Systems Manager doc-only update for June 2023.

# v1.36.6 (2023-06-15)

* No change notes available for this release.

# v1.36.5 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.4 (2023-05-04)

* No change notes available for this release.

# v1.36.3 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.2 (2023-04-10)

* No change notes available for this release.

# v1.36.1 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.36.0 (2023-03-22)

* **Feature**: This Patch Manager release supports creating, updating, and deleting Patch Baselines for AmazonLinux2023, AlmaLinux.

# v1.35.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.
* **Documentation**: Document only update for Feb 2023

# v1.35.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.3 (2023-02-15)

* **Announcement**: When receiving an error response in restJson-based services, an incorrect error type may have been returned based on the content of the response. This has been fixed via PR #2012 tracked in issue #1910.
* **Bug Fix**: Correct error type parsing for restJson services.

# v1.35.2 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.35.1 (2023-01-23)

* No change notes available for this release.

# v1.35.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.34.0 (2023-01-04)

* **Feature**: Adding support for QuickSetup Document Type in Systems Manager

# v1.33.4 (2022-12-21)

* **Documentation**: Doc-only updates for December 2022.

# v1.33.3 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2022-11-22)

* No change notes available for this release.

# v1.33.0 (2022-11-16)

* **Feature**: This release adds support for cross account access in CreateOpsItem, UpdateOpsItem and GetOpsItem. It introduces new APIs to setup resource policies for SSM resources: PutResourcePolicy, GetResourcePolicies and DeleteResourcePolicy.

# v1.32.1 (2022-11-10)

* No change notes available for this release.

# v1.32.0 (2022-11-07)

* **Feature**: This release includes support for applying a CloudWatch alarm to multi account multi region Systems Manager Automation

# v1.31.3 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2022-10-20)

* No change notes available for this release.

# v1.31.0 (2022-10-13)

* **Feature**: Support of AmazonLinux2022 by Patch Manager

# v1.30.0 (2022-09-26)

* **Feature**: This release includes support for applying a CloudWatch alarm to Systems Manager capabilities like Automation, Run Command, State Manager, and Maintenance Windows.

# v1.29.0 (2022-09-23)

* **Feature**: This release adds new SSM document types ConformancePackTemplate and CloudFormation

# v1.28.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.0 (2022-09-14)

* **Feature**: Fixed a bug in the API client generation which caused some operation parameters to be incorrectly generated as value types instead of pointer types. The service API always required these affected parameters to be nilable. This fixes the SDK client to match the expectations of the the service API.
* **Feature**: This release adds support for Systems Manager State Manager Association tagging.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.13 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.12 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.11 (2022-08-30)

* No change notes available for this release.

# v1.27.10 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.9 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.8 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.7 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.6 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.5 (2022-07-27)

* **Documentation**: Adding doc updates for OpsCenter support in Service Setting actions.

# v1.27.4 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.3 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.2 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.1 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2022-05-04)

* **Feature**: This release adds the TargetMaps parameter in SSM State Manager API.

# v1.26.0 (2022-04-29)

* **Feature**: Update the StartChangeRequestExecution, adding TargetMaps to the Runbook parameter

# v1.25.1 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2022-04-19)

* **Feature**: Added offset support for specifying the number of days to wait after the date and time specified by a CRON expression when creating SSM association.

# v1.24.1 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2022-03-25)

* **Feature**: This Patch Manager release supports creating, updating, and deleting Patch Baselines for Rocky Linux OS.

# v1.23.1 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2022-03-23)

* **Feature**: Update AddTagsToResource, ListTagsForResource, and RemoveTagsFromResource APIs to reflect the support for tagging Automation resources. Includes other minor documentation updates.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.0 (2022-01-14)

* **Feature**: Updated API models
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: Updated to latest service endpoints

# v1.17.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2021-11-30)

* **Feature**: API client updated

# v1.16.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.
* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.14.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-10-21)

* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-09-24)

* **Feature**: API client updated

# v1.10.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-08-12)

* **Feature**: API client updated

# v1.8.1 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-07-15)

* **Feature**: Updated service model to latest version.
* **Documentation**: Updated service model to latest revision.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-06-25)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.2 (2021-06-04)

* **Documentation**: Updated service client to latest API model.

# v1.6.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Feature**: Updated to latest service API model.
* **Dependency Update**: Updated to the latest SDK module versions

