// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package extensionerrors

import (
	"fmt"
	"github.com/pkg/errors"
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

func CombineErrors(err1 error, err2 error) error {
	if err1 == nil && err2 == nil {
		return nil
	}
	if err1 != nil && err2 == nil {
		return err1
	}
	if err1 == nil && err2 != nil {
		return err2
	}
	return errors.Wrap(err1, err2.Error())
}
