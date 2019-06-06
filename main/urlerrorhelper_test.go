package main

import (
	"fmt"
	"strings"
	"testing"
)

func Test_urlparse01(t *testing.T) {
	errorString :=	`Get https://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r: dial tcp 13.68.165.64:443: i/o timeout`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://"){
		t.Error("Url removal failed")
	} else {
		fmt.Print(outputErr.Error())
	}
}

func Test_urlparse02(t *testing.T) {
	errorString :=	`Gethttps://aaaaaaaa.blob.core.windows.net/nodeagentpackage-version9-0-0-381/Ubuntu-16.04/batch_config-ubuntu-16.04-1.5.9.tar.gz?sv=2018-03-28&sr=b&sig=a%secret%2Bsecret&st=2019-05-17T01%3A25%3A42Z&se=2021-05-24T01%3A25%3A42Z&sp=r: dial tcp 13.68.165.64:443: i/o timeout`
	inputErr := fmt.Errorf("%s", errorString)
	outputErr := RemoveUrlFromErr(inputErr)
	if strings.Contains(outputErr.Error(), "https://"){
		t.Error("Url removal failed")
	} else {
		fmt.Print(outputErr.Error())
	}
}


