package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/koralski/run-command-extension-linux/pkg/blobutil"
	"github.com/koralski/run-command-extension-linux/pkg/download"
	"github.com/koralski/run-command-extension-linux/pkg/preprocess"
	"github.com/pkg/errors"
	"os"
)

// downloadAndProcessURL downloads using the specified downloader and saves it to the
// specified existing directory, which must be the path to the saved file. Then
// it post-processes file based on heuristics.
func downloadAndProcessURL(ctx *log.Context, url, downloadDir string, cfg *handlerSettings) error {
	fn, err := urlToFileName(url)
	if err != nil {
		return err
	}

	if !urlutil.IsValidUrl(url) {
		return fmt.Errorf("[REDACTED] is not a valid url")
	}

	dl, err := getDownloaders(url, cfg.StorageAccountName, cfg.StorageAccountKey, cfg.ManagedIdentity)
	if err != nil {
		return err
	}

	fp := filepath.Join(downloadDir, fn)
	const mode = 0500 // we assume users download scripts to execute
	if _, err := download.SaveTo(ctx, dl, fp, mode); err != nil {
		return err
	}

	if cfg.SkipDos2Unix == false {
		err = postProcessFile(fp)
	}
	return errors.Wrapf(err, "failed to post-process '%s'", fn)
}

// getDownloader returns a downloader for the given URL based on whether the
// storage credentials are empty or not.
func getDownloaders(fileURL string, storageAccountName, storageAccountKey string, managedIdentity *clientOrObjectId) (
	[]download.Downloader, error) {
	if storageAccountName == "" || storageAccountKey == "" {
		// storage account name and key cannot be specified with managed identity, handler settings validation won't allow that
		// handler settings validation will also not allow storageAccountName XOR storageAccountKey == 1
		// in this case, we can be sure that storage account name and key was not specified
		if download.IsAzureStorageBlobUri(fileURL) && managedIdentity != nil {
			// if managed identity was specified in the configuration, try to use it to download the files
			var msiProvider download.MsiProvider
			switch {
			case managedIdentity.ClientId == "" && managedIdentity.ObjectId == "":
				// get msi using clientId or objectId or implicitly
				msiProvider = download.GetMsiProviderForStorageAccountsImplicitly(fileURL)
			case managedIdentity.ClientId != "" && managedIdentity.ObjectId == "":
				msiProvider = download.GetMsiProviderForStorageAccountsWithClientId(fileURL, managedIdentity.ClientId)
			case managedIdentity.ClientId == "" && managedIdentity.ObjectId != "":
				msiProvider = download.GetMsiProviderForStorageAccountsWithObjectId(fileURL, managedIdentity.ObjectId)
			default:
				return nil, fmt.Errorf("unexpected combination of ClientId and ObjectId found")
			}
			return []download.Downloader{
				// try downloading without MSI token first, but attempt with MSI if the download fails
				download.NewURLDownload(fileURL),
				download.NewBlobWithMsiDownload(fileURL, msiProvider),
			}, nil
		} else {
			// do not use MSI downloader if the uri is not azure storage blob, or managedIdentity isn't specified
			return []download.Downloader{download.NewURLDownload(fileURL)}, nil
		}
	} else {
		// if storage name account and key are specified, use that for all files
		// this preserves old behavior
		blob, err := blobutil.ParseBlobURL(fileURL)
		if err != nil {
			return nil, err
		}
		return []download.Downloader{download.NewBlobDownload(
			storageAccountName, storageAccountKey,
			blob)}, nil
	}
}

// urlToFileName parses given URL and returns the section after the last slash
// character of the path segment to be used as a file name. If a value is not
// found, an error is returned.
func urlToFileName(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse URL: %q", fileURL)
	}

	s := strings.Split(u.Path, "/")
	if len(s) > 0 {
		fn := s[len(s)-1]
		if fn != "" {
			return fn, nil
		}
	}
	return "", fmt.Errorf("cannot extract file name from URL: %q", fileURL)
}

// postProcessFile determines if path is a script file based on heuristics
// and makes in-place changes to the file with some post-processing such as BOM
// and DOS-line endings fixes to make the script POSIX-friendly.
func postProcessFile(path string) error {
	ok, err := preprocess.IsTextFile(path)
	if err != nil {
		return errors.Wrapf(err, "error determining if script is a text file")
	}
	if !ok {
		return nil
	}

	b, err := ioutil.ReadFile(path) // read the file into memory for processing
	if err != nil {
		return errors.Wrapf(err, "error reading file")
	}
	b = preprocess.RemoveBOM(b)
	b = preprocess.Dos2Unix(b)

	err = ioutil.WriteFile(path, b, 0)
	return errors.Wrap(os.Rename(path, path), "error writing file")
}
