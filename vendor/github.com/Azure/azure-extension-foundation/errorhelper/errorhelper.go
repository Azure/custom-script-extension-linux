package errorhelper

import (
	"fmt"
	"runtime/debug"
)

func AddStackToError(err error) error {
	if err == nil {
		return nil
	}
	stackString := string(debug.Stack())
	return fmt.Errorf("%+v\nCallStack: %s", err, stackString)
}

func NewErrorWithStack(errString string) error {
	stackString := string(debug.Stack())
	return fmt.Errorf("%s\nCallStack: %s", errString, stackString)
}
