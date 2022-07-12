// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package lockedfile

import (
	"encoding/json"
	"time"
)

// this is how you do enums in golang
type UpdateMetadataOperation int

const (
	updateOpenTime UpdateMetadataOperation = iota
	updateCloseTime
)

type Metadata struct {
	LastOpened string `json:"LastOpened"`
	LastClosed string `json:"LastClosed"`
}

func (self *Metadata) SetLastOpenedToNow() {
	now := time.Now()
	self.LastOpened = now.Format(time.RFC3339Nano)
}

func (self *Metadata) SetLastClosedToNow() {
	now := time.Now()
	self.LastClosed = now.Format(time.RFC3339Nano)
}

func (self *Metadata) writeMetadataToLockedFile(lockedFile ILockedFile) error {
	bytes, err := json.Marshal(self)
	if err != nil {
		return err
	}
	return lockedFile.WriteLockedFile(bytes)
}

func (self *Metadata) updateAndWriteMetadata(lockedFile ILockedFile, updateOperation UpdateMetadataOperation) error {
	switch updateOperation {
	case updateOpenTime:
		self.SetLastOpenedToNow()
	case updateCloseTime:
		self.SetLastClosedToNow()
	}
	return self.writeMetadataToLockedFile(lockedFile)
}
