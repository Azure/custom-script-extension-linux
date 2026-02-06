package errorutil

import (
	"math"
)

const (
	// System errors
	FileDownload_badRequest   int = -41
	FileDownload_unknownError int = -40

	Imds_internalMsiError int = -30

	Internal_badConfig               int = -21
	Internal_couldNotFindCertificate int = -20

	Os_FailedToDeleteDataDir int = -50
	Os_FailedToOpenStdOut    int = -51
	Os_FailedToOpenStdErr    int = -52

	Storage_internalServerError int = -1
	SystemError                 int = 0 // CRP interprets anything > 0 as user errors

	// User errors
	CommandExecution_failedUnknownError      int = 1
	CommandExecution_failureExitCode         int = 2
	CommandExecution_interruptedByVmShutdown int = 3

	CustomerInput_commandToExecuteSpecifiedInTwoPlaces   int = 20
	CustomerInput_fileUrisSpecifiedInTwoPlaces           int = 22
	CustomerInput_commandToExecuteAndScriptNotSpecified  int = 23
	CustomerInput_fileUriContainsNull                    int = 24
	CustomerInput_invalidFileUris                        int = 25
	CustomerInput_storageCredsAndMIBothSpecified         int = 26
	CustomerInput_clientIdObjectIdBothSpecified          int = 27
	CustomerInput_scriptSpecifiedInTwoPlaces             int = 28
	CustomerInput_commandToExecuteAndScriptBothSpecified int = 29
	CustomerInput_incompleteStorageCreds                 int = 30

	FileDownload_unableToCreateDownloadDirectory int = 50
	FileDownload_sasExpired                      int = 51
	FileDownload_accessDenied                    int = 52
	FileDownload_doesNotExist                    int = 53
	FileDownload_networkingError                 int = 54
	FileDownload_genericError                    int = 55
	FileDownload_exceededTimeout                 int = 56

	Msi_notFound                    int = 70
	Msi_doesNotHaveRightPermissions int = 71
	Msi_GenericRetrievalError       int = 72

	// No Error - used as a placeholder value
	// when representing an "empty" ErrorWithClarification
	// or when the error can be treated without the clarification
	NoError int = math.MaxInt
)
