package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

func RemoveUrlFromErr(err error) error {
	strSegments := strings.Split(err.Error(), " ")
	for i, v := range strSegments {
		if IsValidUrl(v) {
			// we found a url
			strSegments[i] = "[REDACTED]"
		}
	}
	return fmt.Errorf("%s", strings.Join(strSegments, " "))
}

func IsValidUrl(urlstring string) bool {
	// Remove leading and trailing quotes
	noQuotes := strings.Trim(urlstring, `"'`)
	u, parseError := url.Parse(noQuotes)
	if parseError == nil && u.Scheme != "" && u.Host != "" {
		return true
	}
	return false
}
