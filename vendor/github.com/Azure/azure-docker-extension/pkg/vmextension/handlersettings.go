package vmextension

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"os/exec"
)

const (
	settingsFileSuffix = ".settings"
)

type handlerSettingsFile struct {
	RuntimeSettings []struct {
		HandlerSettings handlerSettings `json:"handlerSettings"`
	} `json:"runtimeSettings"`
}

type handlerSettings struct {
	PublicSettings          interface{} `json:"publicSettings"`
	ProtectedSettingsBase64 string      `json:"protectedSettings"`
	SettingsCertThumbprint  string      `json:"protectedSettingsCertThumbprint"`
}

// UnmarshalHandlerSettings locates the latest configuration that should
// be picked up, parses and decodes it by locating the relevant certs
// and returns public and protected settings into the specified instances.
func UnmarshalHandlerSettings(configFolder string, publicSettings, protectedSettings interface{}) error {
	b, err := readSettings(configFolder)
	if err != nil {
		return err
	}
	hs, err := parseHandlerSettingsFile(b)
	if err != nil {
		return err
	}

	// Parse public settings
	if err := unmarshalPublicSettings(hs.PublicSettings, &publicSettings); err != nil {
		return err
	}

	// Parse protected settings
	if err := unmarshalProtectedSettings(configFolder, hs, &protectedSettings); err != nil {
		return err
	}
	return nil
}

// readSettings returns the runtime configuration JSON file with
// the highest sequence number.
func readSettings(configFolder string) ([]byte, error) {
	seq, err := FindSeqNum(configFolder)
	if err != nil {
		return nil, fmt.Errorf("Cannot find seqnum: %v", err)
	}
	cf := filepath.Join(configFolder, fmt.Sprintf("%d%s", seq, settingsFileSuffix))
	b, err := ioutil.ReadFile(cf)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %v", cf, err)
	}
	return b, nil
}

// parseHandlerSettings parses a handler settings file (e.g. 0.settings)
// and returns as an object.
func parseHandlerSettingsFile(b []byte) (h handlerSettings, _ error) {
	if len(b) == 0 { // apparently if no config is specified, we get an empty file
		return h, nil
	}

	var f handlerSettingsFile
	if err := json.Unmarshal(b, &f); err != nil {
		return h, fmt.Errorf("error parsing json: %v", err)
	}
	if len(f.RuntimeSettings) != 1 {
		return h, fmt.Errorf("wrong runtimeSettings count. expected:1, got:%d", len(f.RuntimeSettings))
	}
	return f.RuntimeSettings[0].HandlerSettings, nil
}

// unmarshalPublicSettings parses public settings object serialized
// from handler runtime settings JSON before into the given struct v.
func unmarshalPublicSettings(in interface{}, v interface{}) error {
	s, err := json.Marshal(in)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(s, &v); err != nil {
		return fmt.Errorf("error deserializing public settings for handler: %v", err)
	}
	return nil
}

// unmarshalProtectedSettings decodes the protected settings from
// handler runtime settings JSON file, decrypts it using the certificates
// and unmarshals into the given struct v.
func unmarshalProtectedSettings(configFolder string, hs handlerSettings, v interface{}) error {
	if hs.ProtectedSettingsBase64 == "" {
		return nil
	}
	if hs.SettingsCertThumbprint == "" {
		return errors.New("HandlerSettings has protected settings but no cert thumbprint")
	}

	decoded, err := base64.StdEncoding.DecodeString(hs.ProtectedSettingsBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %v", err)
	}

	// go two levels up where certs are placed (/var/lib/waagent)
	crt := filepath.Join(configFolder, "..", "..", fmt.Sprintf("%s.crt", hs.SettingsCertThumbprint))
	prv := filepath.Join(configFolder, "..", "..", fmt.Sprintf("%s.prv", hs.SettingsCertThumbprint))

	// we use os/exec instead of azure-docker-extension/pkg/executil here as
	// other extension handlers depend on this package for parsing handler
	// settings.
	cmd := exec.Command("openssl", "smime", "-inform", "DER", "-decrypt", "-recip", crt, "-inkey", prv)
	var bOut, bErr bytes.Buffer
	cmd.Stdin = bytes.NewReader(decoded)
	cmd.Stdout = &bOut
	cmd.Stderr = &bErr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("decrypting protected settings failed: error=%v stderr=%v", err, bErr.Bytes())
	}

	// decrypted: json object for protected settings
	if err := json.Unmarshal(bOut.Bytes(), &v); err != nil {
		return fmt.Errorf("failed to unmarshal decrypted settings json: %v", err)
	}
	return nil
}
