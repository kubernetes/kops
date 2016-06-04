package fi

// HasName indicates that the task has a Name
type HasName interface {
	GetName() *string
	SetName(name string)
}
