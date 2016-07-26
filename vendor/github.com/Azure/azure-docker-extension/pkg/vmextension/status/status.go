package status

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type StatusReport []StatusItem

type StatusItem struct {
	Version      float64 `json:"version"`
	TimestampUTC string  `json:"timestampUTC"`
	Status       Status  `json:"status"`
}

type Type string

const (
	StatusTransitioning Type = "transitioning"
	StatusError         Type = "error"
	StatusSuccess       Type = "success"
)

type Status struct {
	Operation        string           `json:"operation"`
	Status           Type             `json:"status"`
	FormattedMessage FormattedMessage `json:"formattedMessage"`
}
type FormattedMessage struct {
	Lang    string `json:"lang"`
	Message string `json:"message"`
}

func NewStatus(t Type, operation, message string) StatusReport {
	return []StatusItem{
		{
			Version:      1.0,
			TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			Status: Status{
				Operation: operation,
				Status:    t,
				FormattedMessage: FormattedMessage{
					Lang:    "en",
					Message: message},
			},
		},
	}
}

func (r StatusReport) marshal() ([]byte, error) {
	return json.MarshalIndent(r, "", "\t")
}

// Save persists the status message to the specified status folder
// using the sequence number.
func (r StatusReport) Save(statusFolder string, seqNum int) error {
	fn := fmt.Sprintf("%d.status", seqNum)
	fp := filepath.Join(statusFolder, fn)

	b, err := r.marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fp, b, 0644)
}
