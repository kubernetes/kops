package fi

// CompareWithID indicates that the value should be compared by the returned ID value (instead of a deep comparison)
// Most Tasks implement this, because typically when a Task references another task, it only is concerned with
// being linked to that task, not the values of the task.
// For example, when an instance is linked to a disk, it cares that the disk is attached to that instance,
// not the size or speed of the disk.
type CompareWithID interface {
	CompareWithID() *string
}
