## Initializers v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | Initializers



Initializers tracks the progress of initialization.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#objectmeta-v1-meta">ObjectMeta meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
pending <br /> *[Initializer](#initializer-v1-meta) array*  <br /> **patch type**: *merge*  <br /> **patch merge key**: *name*  | Pending is a list of initializers that must execute in order before this object is visible. When the last pending initializer is removed, and no failing result is set, the initializers struct will be set to nil and the object is considered as initialized and visible to all clients.
result <br /> *[Status](#status-v1-meta)*    | If result is set with the Failure field, the object will be persisted to storage and then deleted, ensuring that other clients can observe the deletion.

