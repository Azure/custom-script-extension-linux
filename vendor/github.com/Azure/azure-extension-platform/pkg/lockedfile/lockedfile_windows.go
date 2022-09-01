// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package lockedfile

import (
	"golang.org/x/sys/windows"
	"syscall"
	"time"
)

const (
	reserved                    = 0
	allBytes                    = ^uint32(0)
	fileIOTimeoutInMilliseconds = 10000
)

type lockedFile struct {
	fileHandle windows.Handle
	metadata   *Metadata
}

func newInner(filePath string, timeout time.Duration, metadata *Metadata) (*lockedFile, error) {
	name, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, err
	}

	// Open for asynchronous I/O so that we can timeout waiting for the lock.
	// Also open shared so that other processes can open the file (but will
	// still need to lock it).
	handle, err := windows.CreateFile(
		name,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		uint32(windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE),
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_FLAG_OVERLAPPED|windows.FILE_ATTRIBUTE_NORMAL,
		0)

	if err != nil {
		return nil, err
	}

	ol, err := getOverlapped()
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(ol.HEvent)

	err = windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK, reserved, allBytes, allBytes, ol)
	if err == nil {
		return &lockedFile{handle, metadata}, nil
	}

	// ERROR_IO_PENDING is expected when we're waiting on an asynchronous event
	// to occur.
	if err != syscall.ERROR_IO_PENDING {
		return nil, err
	}

	timeoutInMilliseconds := uint32(timeout / time.Millisecond)
	s, err := windows.WaitForSingleObject(ol.HEvent, timeoutInMilliseconds)

	switch s {
	case syscall.WAIT_OBJECT_0:
		// success!
		return &lockedFile{handle, metadata}, nil
	case syscall.WAIT_TIMEOUT:
		windows.CancelIo(handle)
		return nil, &FileLockTimeoutError{"file lock could not be acquired in the specified time"}
	default:
		return nil, err
	}
}

func (self *lockedFile) ReadLockedFile() ([]byte, error) {
	// make an empty byte slice with 4KB default size
	fileBytes := make([]byte, 0, 4096)
	buffer := make([]byte, 4096, 4096)

	ol, err := getOverlapped()
	if err != nil {
		return nil, err
	}
	defer windows.Close(ol.HEvent)
	for {
		err := windows.ReadFile(self.fileHandle, buffer, nil, ol)
		if err != nil && err != syscall.ERROR_IO_PENDING {
			return nil, err
		}
		var readBytes uint32
		err = windows.GetOverlappedResult(self.fileHandle, ol, &readBytes, true)
		if err != nil {
			if err == windows.ERROR_HANDLE_EOF {
				break
			}
			return nil, err
		}

		fileBytes = append(fileBytes, buffer[:readBytes]...)

		// modify ol to read next bytes
		longOffset := combineTwoUint32ToUlong(ol.OffsetHigh, ol.Offset)
		longOffset += uint64(readBytes)
		ol.OffsetHigh, ol.Offset = splitUlongToTwoUint32(longOffset)
		err = windows.ResetEvent(ol.HEvent)
		if err != nil {
			return nil, err
		}
	}
	return fileBytes, nil
}

// warning this function does not resize the file when overwriting to a smaller size
func (self *lockedFile) WriteLockedFile(bytes []byte) error {
	ol, err := getOverlapped()
	if err != nil {
		return err
	}
	defer windows.Close(ol.HEvent)

	err = windows.WriteFile(self.fileHandle, bytes, nil, ol)

	if err != syscall.ERROR_IO_PENDING {
		return err
	}

	s, err := windows.WaitForSingleObject(ol.HEvent, fileIOTimeoutInMilliseconds)

	switch s {
	case syscall.WAIT_OBJECT_0:
		// success writing file
		return err
	case syscall.WAIT_TIMEOUT:
		windows.CancelIo(self.fileHandle)
		return &FileIoTimeout{"fileIO timed out"}
	default:
		return err
	}

	return nil
}

func (self *lockedFile) closeInner() error {
	err := windows.UnlockFileEx(self.fileHandle, reserved, allBytes, allBytes, &windows.Overlapped{HEvent: 0})
	if err != nil {
		return err
	}

	return windows.Close(self.fileHandle)
}

func getOverlapped() (*windows.Overlapped, error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, err
	}

	return &windows.Overlapped{HEvent: event}, nil
}

func splitUlongToTwoUint32(ulong uint64) (high uint32, low uint32) {
	low = uint32(ulong)
	high = uint32(ulong >> 32)
	return
}

func combineTwoUint32ToUlong(high uint32, low uint32) (long uint64) {
	long = uint64(high)
	long = long << 32
	long += uint64(low)
	return long
}
