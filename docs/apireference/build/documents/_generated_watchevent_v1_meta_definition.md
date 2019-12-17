## WatchEvent v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | WatchEvent



Event represents a single event to a watched resource.



Field        | Description
------------ | -----------
object    | Object is:  * If Type is Added or Modified: the new state of the object.  * If Type is Deleted: the state of the object immediately before deletion.  * If Type is Error: *Status is recommended; other types may make sense    depending on context.
type <br /> *string*    | 

