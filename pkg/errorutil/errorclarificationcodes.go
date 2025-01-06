package errorutil

import (
	"math"
)

const (
	// System errors
	fileDownload_badRequest   int = -41
	fileDownload_unknownError int = -40

	imds_internalMsiError int = -30

	internal_badConfig               int = -21
	internal_couldNotFindCertificate int = -20

	storage_internalServerError int = -1
	systemError                 int = 0 // CRP interprets anything > 0 as user errors

	// User errors
	commandExecution_failedUnknownError      int = 1
	commandExecution_failureExitCode         int = 2
	commandExecution_interruptedByVmShutdown int = 3

	customerInput_commandToExecuteSpecifiedInTwoPlaces                   int = 20
	customerInput_ignoreRelativePathForFileDownloadsSpecifiedInTwoPlaces int = 21
	customerInput_fileUrisSpecifiedInTwoPlaces                           int = 22
	customerInput_commandToExecuteNotSpecified                           int = 23
	customerInput_fileUriContainsNull                                    int = 24
	customerInput_invalidFileUris                                        int = 25
	customerInput_storageCredsAndMIBothSpecified                         int = 26
	customerInput_clientIdObjectIdBothSpecified                          int = 27

	fileDownload_unableToCreateDownloadDirectory int = 50
	fileDownload_sasExpired                      int = 51
	fileDownload_accessDenied                    int = 52
	fileDownload_doesNotExist                    int = 53
	fileDownload_networkingError                 int = 54
	fileDownload_genericError                    int = 55
	fileDownload_exceededTimeout                 int = 56

	msi_notFound                    int = 70
	msi_doesNotHaveRightPermissions int = 71
	msi_GenericRetrievalError       int = 72

	// No Error - used as a placeholder value
	// when representing an "empty" ErrorWithClarification
	noError int = math.MaxInt
)
