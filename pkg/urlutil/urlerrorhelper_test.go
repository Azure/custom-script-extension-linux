package urlutil

import (
	"fmt"
	"strings"
	"testing"
)

func Test_expectedErrorString(t *testing.T) {
	errorString := `Get https://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r: dial tcp 13.68.165.64:443: i/o timeout`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://") || strings.Contains(outputErr.Error(), "secret") || !strings.Contains(outputErr.Error(), "[REDACTED]") {
		t.Error("Url removal failed")
	} else if !strings.Contains(outputErr.Error(), "dial tcp 13.68.165.64:443: i/o timeout") {
		t.Error("rest of the error not preserved")
	} else {
		fmt.Println(outputErr.Error())
	}
}

func Test_nonHttpsSchema(t *testing.T) {
	errorString := `Gethttps://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r: dial tcp 13.68.165.64:443: i/o timeout`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://") || strings.Contains(outputErr.Error(), "secret") || !strings.Contains(outputErr.Error(), "[REDACTED]") {
		t.Error("Url removal failed")
	} else if !strings.Contains(outputErr.Error(), "dial tcp 13.68.165.64:443: i/o timeout") {
		t.Error("rest of the error not preserved")
	} else {
		fmt.Println(outputErr.Error())
	}
}

func Test_errorWithMissingSpaces(t *testing.T) {
	errorString := `Gethttps://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r:dial tcp 13.68.165.64:443:i/o timeout`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://") || strings.Contains(outputErr.Error(), "secret") || !strings.Contains(outputErr.Error(), "[REDACTED]") {
		t.Error("Url removal failed")
	} else if !strings.Contains(outputErr.Error(), "tcp 13.68.165.64:443:i/o timeout") {
		t.Error("rest of the error not preserved")
	} else {
		fmt.Println(outputErr.Error())
	}
}

func Test_errorStringIsUrlOnly(t *testing.T) {
	// error string where the uri isn't separated by space with the rest of the error message
	errorString := `https://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://") || strings.Contains(outputErr.Error(), "secret") || !strings.Contains(outputErr.Error(), "[REDACTED]") {
		t.Error("Url removal failed")
	} else {
		fmt.Println(outputErr.Error())
	}
}

func Test_errorStringIsSpaces(t *testing.T) {
	errorString := `   `
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if outputErr.Error() != errorString {
		t.Error("Url detected where there was no URL")
	} else {
		fmt.Println(outputErr.Error())
	}
}

func Test_errorStringHasNoUrl(t *testing.T) {
	errorString := `This: error\\ message \ has \ no? {urls:}`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if outputErr.Error() != errorString {
		t.Error("Url detected where there was no URL")
	} else {
		fmt.Println(outputErr.Error())
	}
}
