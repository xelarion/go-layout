package scheduler

import "errors"

var (
	// ErrTaskAlreadyExists is returned when trying to register a task with a name that already exists.
	ErrTaskAlreadyExists = errors.New("task already exists")
)
