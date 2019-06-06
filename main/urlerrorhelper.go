package main

import (
	"fmt"
	"net/url"
	"strings"
)

func RemoveUrlFromErr(err error) error{
	strSegments := strings.Split(err.Error(), " ")
	for i, v := range strSegments{
		u, parseError := url.Parse(v)
		if parseError == nil && u.Scheme != "" && u.Host != "" && u.Path != ""{
			// we found a url
			strSegments[i] = "[uri redacted]"
		}
	}
	return fmt.Errorf(strings.Join(strSegments, " "))
}
