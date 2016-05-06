package fi

type Task interface {
	Run(*Context) error
}
