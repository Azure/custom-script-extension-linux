// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package extensionerrors

import "github.com/pkg/errors"

var (
	// ErrArgCannotBeNull is returned if a required parameter is null
	ErrArgCannotBeNull = errors.New("The argument cannot be null")

	// ErrArgCannotBeNullOrEmpty is returned if a required string parameter is null or empty
	ErrArgCannotBeNullOrEmpty = errors.New("The argument cannot be null or empty")

	// ErrMustRunAsAdmin is returned if an operation ran at permissions below admin
	ErrMustRunAsAdmin = errors.New("The process must run as Administrator")

	// ErrCertWithThumbprintNotFound is returned if we couldn't find the cert
	ErrCertWithThumbprintNotFound = errors.New("The certificate for the specified thumbprint was not found")

	// ErrInvalidProtectedSettingsData is returned when the protected settings data is invalid
	ErrInvalidProtectedSettingsData = errors.New("The protected settings data is invalid")

	// ErrInvalidSettingsFile is returned if the settings file is invalid
	ErrInvalidSettingsFile = errors.New("The settings file is invalid")

	// ErrInvalidSettingsRuntimeSettingsCount is returned if the runtime settings count is not one
	ErrInvalidSettingsRuntimeSettingsCount = errors.New("The runtime settings count in the settings file is invalid")

	// ErrNoCertificateThumbprint is returned if protected setting exist but no certificate thumbprint does
	ErrNoCertificateThumbprint = errors.New("No certificate thumbprint to decode protected settings")

	// ErrCannotDecodeProtectedSettings is returned if we cannot base64 decode the protected settings
	ErrCannotDecodeProtectedSettings = errors.New("Failed to base64 decode the protected settings")

	// ErrInvalidSettingsFileName is returned if we cannot parse the .settings file name
	ErrInvalidSettingsFileName = errors.New("Invalid .settings file name")

	// ErrNoSettingsFiles is returned if no .settings file are found
	ErrNoSettingsFiles = errors.New("No .settings files exist")

	// ErrNoMrseqFile is returned if no mrseq file are found
	ErrNoMrseqFile = errors.New("No mrseq file exist")

	ErrNotFound = errors.New("NotFound")

	ErrInvalidOperationName = errors.New("operation name is invalid")
)
