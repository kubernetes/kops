package fi

type Target interface {
	// Lifecycle methods, called by the driver
	Finish(taskMap map[string]Task) error
}
