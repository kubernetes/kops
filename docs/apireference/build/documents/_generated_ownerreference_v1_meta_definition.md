## OwnerReference v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | OwnerReference



OwnerReference contains enough information to let you identify an owning object. Currently, an owning object must be in the same namespace, so there is no namespace field.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#objectmeta-v1-meta">ObjectMeta meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
apiVersion <br /> *string*    | API version of the referent.
blockOwnerDeletion <br /> *boolean*    | If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned.
controller <br /> *boolean*    | If true, this reference points to the managing controller.
kind <br /> *string*    | Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
name <br /> *string*    | Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names
uid <br /> *string*    | UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids

