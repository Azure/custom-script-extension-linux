package blobutil

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// AzureBlobRef contains information parsed from an Azure Blob Storage URL.
type AzureBlobRef struct {
	// StorageBase describes the storage endpoint for the blob (without "blob." prefix)
	// e.g. "core.windows.net"
	StorageBase string

	// Container is the container name for the blob
	Container string

	// Blob is the name for the blob (i.e. the rest of the URL after container name) and
	// may contain slashes (/). e.g. "a.txt", "a/b.txt"
	Blob string

	// Scheme contains http or https
	Scheme string
}

// ParseBlobURL recognizes a given Azure Blob Storage URL and extracts the information
// of the blob such as storage API endpoint, container and blob name, or returns error if the
// URL is unrecognized.
func ParseBlobURL(blobURL string) (v AzureBlobRef, err error) {
	u, err := url.Parse(blobURL)
	if err != nil {
		return v, errors.Wrapf(err, "cannot parse URL: %q", blobURL)
	}
	v.Scheme = strings.ToLower(u.Scheme)
	if v.Scheme != "http" && v.Scheme != "https" {
		return v, fmt.Errorf("unsupported scheme in URL: %q", blobURL)
	}

	hostParts := strings.Split(u.Host, ".") // {account}.blob.{storageBase}
	if len(hostParts) < 3 {
		return v, fmt.Errorf("cannot parse azure blob URL: %q", blobURL)
	}
	if strings.ToLower(hostParts[1]) != "blob" {
		return v, fmt.Errorf("blob host not in *.blob.* format: %q", blobURL)
	}

	v.StorageBase = strings.Join(hostParts[2:], ".") // glue them back
	if v.StorageBase == "" {
		return v, fmt.Errorf("cannot parse azure storage endpoint in blob URL: %q", blobURL)
	}

	var ok bool
	v.Container, v.Blob, ok = parseAzureContainerBlobNameFromPath(u.Path)
	if !ok {
		return v, fmt.Errorf("cannot extract Azure container/blob name from: %q", blobURL)
	}

	return v, nil
}

// parseAzureContainerBlobNameFromPath parses Azure Container and blob name from
// path like "/a/b.txt" or "a/b.txt" and can infer the $root container. If it fails
// to parse, returns false.
func parseAzureContainerBlobNameFromPath(path string) (container string, blob string, parsed bool) {
	path = strings.TrimPrefix(path, "/") // remove preceeding /
	if path == "" {
		return "", "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return "$root", parts[0], true
	}
	return parts[0], strings.Join(parts[1:], "/"), true
}
