# Azure Custom Script Virtual Machine Extension

Custom Script Extension lets you run script you provide on Virtual Machines to
bootstrap/install software, run administrative and automation tasks. It can run
an inline script you specify or download a script file from internet or Azure
Storage.

You can add Custom Script Extension to your VM using:

- Azure CLI
- Azure PowerShell
- Azure Resource Manager (ARM) Templates
- Azure Virtual Machines REST API

# User Guide

## 1. Extension Configuration

You can specify files to download and command to execute in the configuration
section of the extensions. If your `commandToExecute` contains secrets, please
use the “protected configuration” section to specify it.

The specified command will be executed only once. If you change anything in the
extension configuration and deploy it again, the command will be executed again.

### 1.1. Public Configuration

Schema for the public configuration file looks like this:

* `commandToExecute`: (**required**, string) the entrypoint script to execute
* `fileUris`: (optional, string array) the URLs for file(s) to be downloaded.
* `timestamp` (optional, integer) use this field only to trigger a re-run of the
  script by changing value of this field.
 
```json
{
  "fileUris": ["<url>"],
  "commandToExecute": "<command-to-execute>"
}
```

> Examples:
>
> ```
> {
>   "fileUris": ["https://gist.github.com/ahmetalpbalkan/b5d4a856fe15464015ae87d5587a4439/raw/466f5c30507c990a4d5a2f5c79f901fa89a80841/hello.sh"],
>   "commandToExecute": "./hello.sh"
> }
> ```
> 
> ```
> {
>   "commandToExecute": "apt-get -y update && apt-get install -y apache2"
> }
> ```


### 1.2. Protected Configuration

The configuration provided in these keys are stored as encrypted and are only
decrypted inside your Virtual Machine:

* `commandToExecute`: (optional, string) the entrypoint script to execute. Use
  this field instead if your command contains secrets such as passwords.
* `storageAccountName`: (optional, string) the name of storage account. If you
  specify storage credentials, all `fileUris` must be URLs for Azure Blobs.
* `storageAccountKey`: (optional, string) the access key of storage account

```json
{
  "commandToExecute": "<command-to-execute>",
  "storageAccountName": "<storage-account-name>",
  "storageAccountKey": "<storage-account-key>"
}
```
 
# 2. Deployment to a Virtual Machine

For **ARM templates**, see [this documentation][doc] to create an extension
resource in your template.

[doc]: https://azure.microsoft.com/documentation/articles/virtual-machines-linux-extensions-customscript/

For **Azure CLI**, create a `public.json` (and optionally `protected.json`) and run:

    $ azure vm extension set <resource-group> <vm-name> \
	    CustomScript Microsoft.Azure.Extensions 2.0 \
	    --auto-upgrade-minor-version \
	    --public-config-path public.json \
	    --private-config-path protected.json



# 3. Troubleshooting

Your files are downloaded to a path like: `/var/lib/azure/custom-script/download/0/` and
the command output is saved to `stdout` and `stderr` files in this directory.

You can find the logs for the extension at `/var/log/azure/custom-script/handler.log`.

Please open an issue on this GitHub repository if you encounter problems that
you could not debug with these log files.  

-----

[![Build Status](https://travis-ci.org/Azure/custom-script-extension-linux.svg?branch=master)](https://travis-ci.org/Azure/custom-script-extension-linux)

-----
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
