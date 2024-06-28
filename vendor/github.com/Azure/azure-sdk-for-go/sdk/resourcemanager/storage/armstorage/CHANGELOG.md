# Release History

## 1.6.0 (2024-06-28)
### Features Added

- New value `AccessTierCold` added to enum type `AccessTier`
- New value `ExpirationActionBlock` added to enum type `ExpirationAction`
- New value `MinimumTLSVersionTLS13` added to enum type `MinimumTLSVersion`
- New value `ProvisioningStateCanceled`, `ProvisioningStateDeleting`, `ProvisioningStateFailed`, `ProvisioningStateValidateSubscriptionQuotaBegin`, `ProvisioningStateValidateSubscriptionQuotaEnd` added to enum type `ProvisioningState`
- New value `PublicNetworkAccessSecuredByPerimeter` added to enum type `PublicNetworkAccess`
- New enum type `IssueType` with values `IssueTypeConfigurationPropagationFailure`, `IssueTypeUnknown`
- New enum type `ListLocalUserIncludeParam` with values `ListLocalUserIncludeParamNfsv3`
- New enum type `NetworkSecurityPerimeterConfigurationProvisioningState` with values `NetworkSecurityPerimeterConfigurationProvisioningStateAccepted`, `NetworkSecurityPerimeterConfigurationProvisioningStateCanceled`, `NetworkSecurityPerimeterConfigurationProvisioningStateDeleting`, `NetworkSecurityPerimeterConfigurationProvisioningStateFailed`, `NetworkSecurityPerimeterConfigurationProvisioningStateSucceeded`
- New enum type `NspAccessRuleDirection` with values `NspAccessRuleDirectionInbound`, `NspAccessRuleDirectionOutbound`
- New enum type `ResourceAssociationAccessMode` with values `ResourceAssociationAccessModeAudit`, `ResourceAssociationAccessModeEnforced`, `ResourceAssociationAccessModeLearning`
- New enum type `RunResult` with values `RunResultFailed`, `RunResultSucceeded`
- New enum type `RunStatusEnum` with values `RunStatusEnumFinished`, `RunStatusEnumInProgress`
- New enum type `Severity` with values `SeverityError`, `SeverityWarning`
- New enum type `TriggerType` with values `TriggerTypeOnSchedule`, `TriggerTypeRunOnce`
- New function `*ClientFactory.NewNetworkSecurityPerimeterConfigurationsClient() *NetworkSecurityPerimeterConfigurationsClient`
- New function `*ClientFactory.NewTaskAssignmentInstancesReportClient() *TaskAssignmentInstancesReportClient`
- New function `*ClientFactory.NewTaskAssignmentsClient() *TaskAssignmentsClient`
- New function `*ClientFactory.NewTaskAssignmentsInstancesReportClient() *TaskAssignmentsInstancesReportClient`
- New function `NewTaskAssignmentInstancesReportClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TaskAssignmentInstancesReportClient, error)`
- New function `*TaskAssignmentInstancesReportClient.NewListPager(string, string, string, *TaskAssignmentInstancesReportClientListOptions) *runtime.Pager[TaskAssignmentInstancesReportClientListResponse]`
- New function `NewTaskAssignmentsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TaskAssignmentsClient, error)`
- New function `*TaskAssignmentsClient.BeginCreate(context.Context, string, string, string, TaskAssignment, *TaskAssignmentsClientBeginCreateOptions) (*runtime.Poller[TaskAssignmentsClientCreateResponse], error)`
- New function `*TaskAssignmentsClient.BeginDelete(context.Context, string, string, string, *TaskAssignmentsClientBeginDeleteOptions) (*runtime.Poller[TaskAssignmentsClientDeleteResponse], error)`
- New function `*TaskAssignmentsClient.Get(context.Context, string, string, string, *TaskAssignmentsClientGetOptions) (TaskAssignmentsClientGetResponse, error)`
- New function `*TaskAssignmentsClient.NewListPager(string, string, *TaskAssignmentsClientListOptions) *runtime.Pager[TaskAssignmentsClientListResponse]`
- New function `*TaskAssignmentsClient.BeginUpdate(context.Context, string, string, string, TaskAssignmentUpdateParameters, *TaskAssignmentsClientBeginUpdateOptions) (*runtime.Poller[TaskAssignmentsClientUpdateResponse], error)`
- New function `NewTaskAssignmentsInstancesReportClient(string, azcore.TokenCredential, *arm.ClientOptions) (*TaskAssignmentsInstancesReportClient, error)`
- New function `*TaskAssignmentsInstancesReportClient.NewListPager(string, string, *TaskAssignmentsInstancesReportClientListOptions) *runtime.Pager[TaskAssignmentsInstancesReportClientListResponse]`
- New function `NewNetworkSecurityPerimeterConfigurationsClient(string, azcore.TokenCredential, *arm.ClientOptions) (*NetworkSecurityPerimeterConfigurationsClient, error)`
- New function `*NetworkSecurityPerimeterConfigurationsClient.Get(context.Context, string, string, string, *NetworkSecurityPerimeterConfigurationsClientGetOptions) (NetworkSecurityPerimeterConfigurationsClientGetResponse, error)`
- New function `*NetworkSecurityPerimeterConfigurationsClient.NewListPager(string, string, *NetworkSecurityPerimeterConfigurationsClientListOptions) *runtime.Pager[NetworkSecurityPerimeterConfigurationsClientListResponse]`
- New function `*NetworkSecurityPerimeterConfigurationsClient.BeginReconcile(context.Context, string, string, string, *NetworkSecurityPerimeterConfigurationsClientBeginReconcileOptions) (*runtime.Poller[NetworkSecurityPerimeterConfigurationsClientReconcileResponse], error)`
- New struct `ExecutionTarget`
- New struct `ExecutionTrigger`
- New struct `ExecutionTriggerUpdate`
- New struct `NetworkSecurityPerimeter`
- New struct `NetworkSecurityPerimeterConfiguration`
- New struct `NetworkSecurityPerimeterConfigurationList`
- New struct `NetworkSecurityPerimeterConfigurationProperties`
- New struct `NetworkSecurityPerimeterConfigurationPropertiesProfile`
- New struct `NetworkSecurityPerimeterConfigurationPropertiesResourceAssociation`
- New struct `NspAccessRule`
- New struct `NspAccessRuleProperties`
- New struct `NspAccessRulePropertiesSubscriptionsItem`
- New struct `ProvisioningIssue`
- New struct `ProvisioningIssueProperties`
- New struct `ProxyResourceAutoGenerated`
- New struct `ResourceAutoGenerated`
- New struct `TaskAssignment`
- New struct `TaskAssignmentExecutionContext`
- New struct `TaskAssignmentProperties`
- New struct `TaskAssignmentReport`
- New struct `TaskAssignmentUpdateExecutionContext`
- New struct `TaskAssignmentUpdateParameters`
- New struct `TaskAssignmentUpdateProperties`
- New struct `TaskAssignmentUpdateReport`
- New struct `TaskAssignmentsList`
- New struct `TaskReportInstance`
- New struct `TaskReportProperties`
- New struct `TaskReportSummary`
- New struct `TriggerParameters`
- New struct `TriggerParametersUpdate`
- New field `EnableExtendedGroups` in struct `AccountProperties`
- New field `EnableExtendedGroups` in struct `AccountPropertiesCreateParameters`
- New field `EnableExtendedGroups` in struct `AccountPropertiesUpdateParameters`
- New field `AllowACLAuthorization`, `ExtendedGroups`, `GroupID`, `IsNFSv3Enabled`, `UserID` in struct `LocalUserProperties`
- New field `NextLink` in struct `LocalUsers`
- New field `Filter`, `Include`, `Maxpagesize` in struct `LocalUsersClientListOptions`


## 1.5.0 (2023-11-24)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.


## 1.5.0-beta.1 (2023-10-09)
### Features Added

- Support for test fakes and OpenTelemetry trace spans.

## 1.4.0 (2023-08-25)
### Features Added

- New value `CorsRuleAllowedMethodsItemCONNECT`, `CorsRuleAllowedMethodsItemTRACE` added to enum type `CorsRuleAllowedMethodsItem`
- New enum type `MigrationName` with values `MigrationNameDefault`
- New enum type `MigrationStatus` with values `MigrationStatusComplete`, `MigrationStatusFailed`, `MigrationStatusInProgress`, `MigrationStatusInvalid`, `MigrationStatusSubmittedForConversion`
- New enum type `PostFailoverRedundancy` with values `PostFailoverRedundancyStandardLRS`, `PostFailoverRedundancyStandardZRS`
- New enum type `PostPlannedFailoverRedundancy` with values `PostPlannedFailoverRedundancyStandardGRS`, `PostPlannedFailoverRedundancyStandardGZRS`, `PostPlannedFailoverRedundancyStandardRAGRS`, `PostPlannedFailoverRedundancyStandardRAGZRS`
- New function `*AccountsClient.BeginCustomerInitiatedMigration(context.Context, string, string, AccountMigration, *AccountsClientBeginCustomerInitiatedMigrationOptions) (*runtime.Poller[AccountsClientCustomerInitiatedMigrationResponse], error)`
- New function `*AccountsClient.GetCustomerInitiatedMigration(context.Context, string, string, MigrationName, *AccountsClientGetCustomerInitiatedMigrationOptions) (AccountsClientGetCustomerInitiatedMigrationResponse, error)`
- New struct `AccountMigration`
- New struct `AccountMigrationProperties`
- New struct `BlobInventoryCreationTime`
- New struct `ErrorAdditionalInfo`
- New struct `ErrorDetail`
- New struct `ErrorResponseAutoGenerated`
- New field `AccountMigrationInProgress`, `IsSKUConversionBlocked` in struct `AccountProperties`
- New field `CreationTime` in struct `BlobInventoryPolicyFilter`
- New field `CanPlannedFailover`, `PostFailoverRedundancy`, `PostPlannedFailoverRedundancy` in struct `GeoReplicationStats`


## 1.3.0 (2023-03-27)
### Features Added

- New struct `ClientFactory` which is a client factory used to create any client in this module

## 1.2.0 (2022-12-23)
### Features Added

- New type alias `ListEncryptionScopesInclude`
- New field `FailoverType` in struct `AccountsClientBeginFailoverOptions`
- New field `TierToCold` in struct `ManagementPolicyBaseBlob`
- New field `TierToHot` in struct `ManagementPolicyBaseBlob`
- New field `Filter` in struct `EncryptionScopesClientListOptions`
- New field `Include` in struct `EncryptionScopesClientListOptions`
- New field `Maxpagesize` in struct `EncryptionScopesClientListOptions`
- New field `TierToHot` in struct `ManagementPolicyVersion`
- New field `TierToCold` in struct `ManagementPolicyVersion`
- New field `TierToCold` in struct `ManagementPolicySnapShot`
- New field `TierToHot` in struct `ManagementPolicySnapShot`


## 1.1.0 (2022-08-10)
### Features Added

- New const `DirectoryServiceOptionsAADKERB`


## 1.0.0 (2022-05-16)

The package of `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage` is using our [next generation design principles](https://azure.github.io/azure-sdk/general_introduction.html) since version 1.0.0, which contains breaking changes.

To migrate the existing applications to the latest version, please refer to [Migration Guide](https://aka.ms/azsdk/go/mgmt/migration).

To learn more, please refer to our documentation [Quick Start](https://aka.ms/azsdk/go/mgmt).