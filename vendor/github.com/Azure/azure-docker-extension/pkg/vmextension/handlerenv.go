package vmextension

import (
	"encoding/json"
	"fmt"
)

type HandlerEnvironment struct {
	Version            float64 `json:"version"`
	SeqNo              string  `json:"seqNo"`
	Name               string  `json:"name"`
	HandlerEnvironment struct {
		HeartbeatFile string `json:"heartbeatFile"`
		StatusFolder  string `json:"statusFolder"`
		ConfigFolder  string `json:"configFolder"`
		LogFolder     string `json:"logFolder"`
	}
}

type HandlerEnvironmentFile []HandlerEnvironment

// ParseHandlerEnv parses the /var/lib/waagent/[extension]/HandlerEnvironment.json
// format.
func ParseHandlerEnv(b []byte) (HandlerEnvironment, error) {
	var hf HandlerEnvironmentFile
	var he HandlerEnvironment
	if err := json.Unmarshal(b, &hf); err != nil {
		return he, fmt.Errorf("failed to parse handler env: %v", err)
	}
	if len(hf) != 1 {
		return he, fmt.Errorf("expected 1 config in HandlerEnvironment, found: %v", len(hf))
	}
	return hf[0], nil
}
