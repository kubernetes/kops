package fi

import (
	"fmt"
)

func RequiredField(key string) error {
	return fmt.Errorf("Field is required: %s", key)
}

func CannotChangeField(key string) error {
	return fmt.Errorf("Field cannot be changed: %s", key)
}
