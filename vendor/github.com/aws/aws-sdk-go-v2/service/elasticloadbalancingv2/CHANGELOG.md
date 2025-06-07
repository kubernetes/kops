# v1.45.2 (2025-04-10)

* No change notes available for this release.

# v1.45.1 (2025-04-03)

* No change notes available for this release.

# v1.45.0 (2025-03-07)

* **Feature**: This release adds support for assigning IP addresses to Application Load Balancers from VPC IP Address Manager pools.

# v1.44.1 (2025-03-04.2)

* **Bug Fix**: Add assurance test for operation order.

# v1.44.0 (2025-02-27)

* **Feature**: Track credential providers via User-Agent Feature ids
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.13 (2025-02-18)

* **Bug Fix**: Bump go version to 1.22
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.12 (2025-02-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.11 (2025-02-04)

* No change notes available for this release.

# v1.43.10 (2025-01-31)

* **Dependency Update**: Switch to code-generated waiter matchers, removing the dependency on go-jmespath.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.9 (2025-01-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.8 (2025-01-24)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade to smithy-go v1.22.2.

# v1.43.7 (2025-01-17)

* **Bug Fix**: Fix bug where credentials weren't refreshed during retry loop.

# v1.43.6 (2025-01-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.5 (2025-01-14)

* **Bug Fix**: Fix issue where waiters were not failing on unmatched errors as they should. This may have breaking behavioral changes for users in fringe cases. See [this announcement](https://github.com/aws/aws-sdk-go-v2/discussions/2954) for more information.

# v1.43.4 (2025-01-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.3 (2025-01-08)

* No change notes available for this release.

# v1.43.2 (2024-12-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.1 (2024-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.43.0 (2024-11-22)

* **Feature**: This release adds support for advertising trusted CA certificate names in associated trust stores.

# v1.42.1 (2024-11-21)

* **Documentation**: This feature adds support for enabling zonal shift on cross-zone enabled Application Load Balancer, as well as modifying HTTP request and response headers.

# v1.42.0 (2024-11-20)

* **Feature**: This release adds support for configuring Load balancer Capacity Unit reservations

# v1.41.2 (2024-11-18)

* **Dependency Update**: Update to smithy-go v1.22.1.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.1 (2024-11-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.41.0 (2024-10-31)

* **Feature**: Add UDP support for AWS PrivateLink and dual-stack Network Load Balancers

# v1.40.1 (2024-10-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.40.0 (2024-10-11)

* **Feature**: Add zonal_shift.config.enabled attribute. Add new AdministrativeOverride construct in the describe-target-health API response to include information about the override status applied to a target.

# v1.39.2 (2024-10-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.1 (2024-10-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.39.0 (2024-10-04)

* **Feature**: Add support for HTTP client metrics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.38.4 (2024-10-03)

* No change notes available for this release.

# v1.38.3 (2024-09-27)

* No change notes available for this release.

# v1.38.2 (2024-09-25)

* No change notes available for this release.

# v1.38.1 (2024-09-23)

* No change notes available for this release.

# v1.38.0 (2024-09-20)

* **Feature**: Add tracing and metrics support to service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.37.1 (2024-09-17)

* **Bug Fix**: **BREAKFIX**: Only generate AccountIDEndpointMode config for services that use it. This is a compiler break, but removes no actual functionality, as no services currently use the account ID in endpoint resolution.

# v1.37.0 (2024-09-12)

* **Feature**: Correct incorrectly mapped error in ELBv2 waiters

# v1.36.0 (2024-09-09)

* **Feature**: Add paginators for the ELBv2 DescribeListenerCertificates and DescribeRules APIs. Fix broken waiter for the ELBv2 DescribeLoadBalancers API.

# v1.35.1 (2024-09-04)

* No change notes available for this release.

# v1.35.0 (2024-09-03)

* **Feature**: This release adds support for configuring TCP idle timeout on NLB and GWLB listeners.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.2 (2024-08-22)

* No change notes available for this release.

# v1.34.1 (2024-08-15)

* **Dependency Update**: Bump minimum Go version to 1.21.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.34.0 (2024-07-25)

* **Feature**: This release adds support for sharing trust stores across accounts and organizations through integration with AWS Resource Access Manager.

# v1.33.3 (2024-07-10.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.2 (2024-07-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.1 (2024-06-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.33.0 (2024-06-26)

* **Feature**: Support list-of-string endpoint parameter.

# v1.32.1 (2024-06-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.32.0 (2024-06-18)

* **Feature**: Track usage of various AWS SDK features in user-agent string.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.4 (2024-06-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.3 (2024-06-07)

* **Bug Fix**: Add clock skew correction on all service clients
* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.2 (2024-06-03)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.31.1 (2024-05-23)

* No change notes available for this release.

# v1.31.0 (2024-05-17)

* **Feature**: This release adds dualstack-without-public-ipv4 IP address type for ALB.

# v1.30.8 (2024-05-16)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.7 (2024-05-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.6 (2024-05-08)

* **Bug Fix**: GoDoc improvement

# v1.30.5 (2024-03-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.4 (2024-03-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.3 (2024-03-14)

* **Documentation**: This release allows you to configure HTTP client keep-alive duration for communication between clients and Application Load Balancers.

# v1.30.2 (2024-03-07)

* **Bug Fix**: Remove dependency on go-cmp.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.30.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.29.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.29.0 (2024-02-16)

* **Feature**: Add new ClientOptions field to waiter config which allows you to extend the config for operation calls made by waiters.

# v1.28.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.27.0 (2024-01-31)

* **Feature**: This release enables unhealthy target draining intervals for Network Load Balancers.

# v1.26.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.6 (2023-12-20)

* No change notes available for this release.

# v1.26.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.26.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.26.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.25.0 (2023-11-27)

* **Feature**: This release enables both mutual authentication (mTLS), and Automatic Target Weights (ATW) for Application Load Balancers.

# v1.24.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2023-10-24)

* **Feature**: **BREAKFIX**: Correct nullability and default value representation of various input fields across a large number of services. Calling code that references one or more of the affected fields will need to update usage accordingly. See [2162](https://github.com/aws/aws-sdk-go-v2/issues/2162).

# v1.21.6 (2023-10-12)

* **Documentation**: This release enables routing policies with Availability Zone affinity for Network Load Balancers.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.5 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2023-09-06)

* **Documentation**: This release enables default UDP connection termination and disabling unhealthy target connection termination for Network Load Balancers.

# v1.21.3 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.0 (2023-08-10)

* **Feature**: This release enables configuring security groups for Network Load Balancers

# v1.20.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.1 (2023-08-01)

* No change notes available for this release.

# v1.20.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.15 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.14 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.13 (2023-06-15)

* No change notes available for this release.

# v1.19.12 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.11 (2023-05-04)

* No change notes available for this release.

# v1.19.10 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.9 (2023-04-10)

* No change notes available for this release.

# v1.19.8 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.7 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.6 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.5 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.19.4 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.19.3 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade smithy to 1.27.2 and correct empty query list serialization.

# v1.19.2 (2023-02-02)

* **Documentation**: The GWLB Flex Health Check project updates the default values of healthy-threshold-count from 3 to 5 and unhealthy-threshold-count from 3 to 2

# v1.19.1 (2023-01-23)

* No change notes available for this release.

# v1.19.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.18.28 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.27 (2022-12-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.26 (2022-11-22)

* No change notes available for this release.

# v1.18.25 (2022-11-17)

* **Documentation**: Provides new target group attributes to turn on/off cross zone load balancing and configure target group health for Network Load Balancers and Application Load Balancers. Provides improvements to health check configuration for Network Load Balancers.

# v1.18.24 (2022-11-16)

* No change notes available for this release.

# v1.18.23 (2022-11-10)

* No change notes available for this release.

# v1.18.22 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.21 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.20 (2022-10-07)

* **Documentation**: Gateway Load Balancer adds a new feature (target_failover) for customers to rebalance existing flows to a healthy target after marked unhealthy or deregistered. This allows graceful patching/upgrades of target appliances during maintenance windows, and helps reduce unhealthy target failover time.

# v1.18.19 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.18 (2022-09-14)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.17 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.16 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.15 (2022-08-30)

* No change notes available for this release.

# v1.18.14 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.13 (2022-08-25)

* **Documentation**: Documentation updates for ELBv2.  Gateway Load Balancer now supports Configurable Flow Stickiness, enabling you to configure the hashing used to maintain stickiness of flows to a specific target appliance.

# v1.18.12 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.11 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.10 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.9 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.8 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.7 (2022-06-29)

* **Documentation**: This release adds two attributes for ALB. One, helps to preserve the host header and the other helps to modify, preserve, or remove the X-Forwarded-For header in the HTTP request.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.6 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.5 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.4 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.3 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: Updated to latest service endpoints

# v1.13.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-11-30)

* **Feature**: API client updated

# v1.12.1 (2021-11-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.
* **Feature**: Waiters now have a `WaitForOutput` method, which can be used to retrieve the output of the successful wait operation. Thank you to [Andrew Haines](https://github.com/haines) for contributing this feature.

# v1.11.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.10.0 (2021-10-21)

* **Feature**: API client updated
* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-10-11)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-09-30)

* **Feature**: API client updated

# v1.7.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.1 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-08-04)

* **Feature**: Updated to latest API model.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-07-15)

* **Feature**: The ErrorCode method on generated service error types has been corrected to match the API model.
* **Documentation**: Updated service model to latest revision.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.3.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

