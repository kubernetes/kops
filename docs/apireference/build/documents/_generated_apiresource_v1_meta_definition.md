## APIResource v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | APIResource



APIResource specifies the name of a resource and whether it is namespaced.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#apiresourcelist-v1-meta">APIResourceList meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
categories <br /> *string array*    | categories is a list of the grouped resources this resource belongs to (e.g. 'all')
group <br /> *string*    | group is the preferred group of the resource.  Empty implies the group of the containing resource list. For subresources, this may have a different value, for example: Scale".
kind <br /> *string*    | kind is the kind for the resource (e.g. 'Foo' is the kind for a resource 'foo')
name <br /> *string*    | name is the plural name of the resource.
namespaced <br /> *boolean*    | namespaced indicates if a resource is namespaced or not.
shortNames <br /> *string array*    | shortNames is a list of suggested short names of the resource.
singularName <br /> *string*    | singularName is the singular name of the resource.  This allows clients to handle plural and singular opaquely. The singularName is more correct for reporting status on a single item and both singular and plural are allowed from the kubectl CLI interface.
verbs <br /> *string array*    | verbs is a list of supported kube verbs (this includes get, list, watch, create, update, patch, delete, deletecollection, and proxy)
version <br /> *string*    | version is the preferred version of the resource.  Empty implies the version of the containing resource list For subresources, this may have a different value, for example: v1 (while inside a v1beta1 version of the core resource's group)".

