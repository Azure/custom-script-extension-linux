// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package logging

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-extension-platform/pkg/handlerenv"
)

const (
	logLevelError   = "Error "
	logLevelWarning = "Warning "
	logLevelInfo    = "Info "
)

const (
	thirtyMB            = 30 * 1024 * 1034 // 31,457,280 bytes
	fortyMB             = 40 * 1024 * 1024 // 41,943,040 bytes
	logDirThresholdLow  = thirtyMB
	logDirThresholdHigh = fortyMB
)

// ExtensionLogger exposes logging capabilities to the extension
// It automatically appends time stamps and debug level to each message
// and ensures all logs are placed in the logs folder passed by the agent
type ExtensionLogger struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	file        *os.File
}

// New creates a new logging instance. If the handlerEnvironment is nil, we'll use a
// standard output logger
func New(he *handlerenv.HandlerEnvironment) *ExtensionLogger {
	return NewWithName(he, "")
}

// Allows the caller to specify their own name for the file
// Supports cycling of logs to prevent filling up the disk
func NewWithName(he *handlerenv.HandlerEnvironment, logFileFormat string) *ExtensionLogger {
	if he == nil {
		return newStandardOutput()
	}

	if logFileFormat == "" {
		logFileFormat = "log_%v"
	}

	// Rotate log folder to prevent filling up the disk
	err := rotateLogFolder(he.LogFolder, logFileFormat)
	if err != nil {
		return newStandardOutput()
	}

	fileName := fmt.Sprintf(logFileFormat, strconv.FormatInt(time.Now().UTC().Unix(), 10))
	filePath := path.Join(he.LogFolder, fileName)
	writer, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return newStandardOutput()
	}

	return &ExtensionLogger{
		errorLogger: log.New(writer, logLevelError, log.Ldate|log.Ltime|log.LUTC),
		infoLogger:  log.New(writer, logLevelInfo, log.Ldate|log.Ltime|log.LUTC),
		warnLogger:  log.New(writer, logLevelWarning, log.Ldate|log.Ltime|log.LUTC),
		file:        writer,
	}
}

func GetCallStack() string {
	return string(debug.Stack())
}

func newStandardOutput() *ExtensionLogger {
	return &ExtensionLogger{
		errorLogger: log.New(os.Stdout, logLevelError, 0),
		infoLogger:  log.New(os.Stdout, logLevelInfo, 0),
		warnLogger:  log.New(os.Stdout, logLevelWarning, 0),
		file:        nil,
	}
}

// Close closes the file
func (logger *ExtensionLogger) Close() {
	if logger.file != nil {
		logger.file.Close()
	}
}

// Error logs an error. Format is the same as fmt.Print
func (logger *ExtensionLogger) Error(format string, v ...interface{}) {
	logger.errorLogger.Printf(format+"\n", v...)
	logger.errorLogger.Printf(GetCallStack() + "\n")
}

// Warn logs a warning. Format is the same as fmt.Print
func (logger *ExtensionLogger) Warn(format string, v ...interface{}) {
	logger.warnLogger.Printf(format+"\n", v...)
}

// Info logs an information statement. Format is the same as fmt.Print
func (logger *ExtensionLogger) Info(format string, v ...interface{}) {
	logger.infoLogger.Printf(format+"\n", v...)
}

// Error logs an error. Get the message from a stream directly
func (logger *ExtensionLogger) ErrorFromStream(prefix string, streamReader io.Reader) {
	logger.errorLogger.Print(prefix)
	io.Copy(logger.errorLogger.Writer(), streamReader)
	logger.errorLogger.Writer().Write([]byte(fmt.Sprintln())) // add a newline at the end of the stream contents
}

// Warn logs a warning. Get the message from a stream directly
func (logger *ExtensionLogger) WarnFromStream(prefix string, streamReader io.Reader) {
	logger.warnLogger.Print(prefix)
	io.Copy(logger.warnLogger.Writer(), streamReader)
	logger.warnLogger.Writer().Write([]byte(fmt.Sprintln())) // add a newline at the end of the stream contents
}

// Info logs an information statement. Get the message from a stream directly
func (logger *ExtensionLogger) InfoFromStream(prefix string, streamReader io.Reader) {
	logger.infoLogger.Print(prefix)
	io.Copy(logger.infoLogger.Writer(), streamReader)
	logger.infoLogger.Writer().Write([]byte(fmt.Sprintln())) // add a newline at the end of the stream contents
}

// Function to get directory size
func getDirSize(dirPath string) (size int64, err error) {
	err = filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	if err != nil {
		err = fmt.Errorf("unable to compute directory size, error: %v", err)
	}
	return
}

// Function to rotate log files present in logFolder to avoid filling customer disk space
// File name matching is done on file name pattern provided before '%'
func rotateLogFolder(logFolder string, logFileFormat string) (err error) {
	size, err := getDirSize(logFolder)
	if err != nil {
		return
	}

	// If directory size is still under high threshold value, nothing to do
	if size < logDirThresholdHigh {
		return
	}

	// Get all log files in logFolder
	// Files are already sorted according to filenames
	// Log file names contains unix timestamp as suffix, Thus we have files sorted according to age as well
	var dirEntries []fs.FileInfo

	dirEntries, err = ioutil.ReadDir(logFolder)
	if err != nil {
		err = fmt.Errorf("unable to read log folder, error: %v", err)
		return
	}

	// Sort directory entries according to time (oldest to newest)
	sort.Slice(dirEntries, func(idx1, idx2 int) bool {
		return dirEntries[idx1].ModTime().Before(dirEntries[idx2].ModTime())
	})

	// Get log file name prefix
	logFilePrefix := strings.Split(logFileFormat, "%")

	for _, file := range dirEntries {
		// Once directory size goes below lower threshold limit, stop deletion
		if size < logDirThresholdLow {
			break
		}

		// Skip directories
		if file.IsDir() {
			continue
		}

		// log file names are prefixed according to logFileFormat specified
		if !strings.HasPrefix(file.Name(), logFilePrefix[0]) {
			continue
		}

		// Delete the file
		err = os.Remove(filepath.Join(logFolder, file.Name()))
		if err != nil {
			err = fmt.Errorf("unable to delete log files, error: %v", err)
			return
		}

		// Subtract file size from total directory size
		size = size - file.Size()
	}
	return
}
