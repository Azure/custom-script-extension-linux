// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package lockedfile

import (
	"time"
)

type ILockedFile interface {
	ReadLockedFile() ([]byte, error)
	WriteLockedFile(bytes []byte) error
	Close() error
}

func New(filePath string, timeout time.Duration) (lockedFile ILockedFile, err error) {
	metadata := Metadata{}
	lockedFile, err = newInner(filePath, timeout, &metadata)
	if err != nil {
		return
	}
	err = metadata.updateAndWriteMetadata(lockedFile, updateOpenTime)
	return
}

func (self *lockedFile) Close() error {
	self.metadata.updateAndWriteMetadata(self, updateCloseTime)
	return self.closeInner()
}
