// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package lockedfile

type FileLockTimeoutError struct {
	message string
}

func (self *FileLockTimeoutError) Error() string {
	return self.message
}

type FileLockGenericError struct {
	message string
}

func (self *FileLockGenericError) Error() string {
	return self.message
}

type FileIoTimeout struct {
	message string
}

func (self *FileIoTimeout) Error() string {
	return self.message
}
