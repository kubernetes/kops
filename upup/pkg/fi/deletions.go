package fi

type ProducesDeletions interface {
	FindDeletions(*Context) ([]Deletion, error)
}

type Deletion interface {
	Delete(target Target) error

	TaskName() string
	Item() string
}
