## StatusDetails v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | StatusDetails



StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#status-v1-meta">Status meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
causes <br /> *[StatusCause](#statuscause-v1-meta) array*    | The Causes array includes more details associated with the StatusReason failure. Not all StatusReasons may provide detailed causes.
group <br /> *string*    | The group attribute of the resource associated with the status StatusReason.
kind <br /> *string*    | The kind attribute of the resource associated with the status StatusReason. On some operations may differ from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
name <br /> *string*    | The name attribute of the resource associated with the status StatusReason (when there is a single name which can be described).
retryAfterSeconds <br /> *integer*    | If specified, the time in seconds before the operation should be retried. Some errors may indicate the client must take an alternate action - for those errors this field may indicate how long to wait before taking the alternate action.
uid <br /> *string*    | UID of the resource. (when there is a single resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids

