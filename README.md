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
> 

Example (PowerShell with Variables):
```
$ConfigScriptURI="https://gist.github.com/ahmetalpbalkan/b5d4a856fe15464015ae87d5587a4439/raw/466f5c30507c990a4d5a2f5c79f901fa89a80841/hello.sh"
$ConfigScriptFileName = "hello.sh"
$Command2Exec = "sh $ConfigScriptFileName"
$PublicConf = '{
   "fileUris": ['+$ConfigScriptUri+'],
   "commandToExecute": "'+$Command2Exec+'"
}'
```

### 1.2. Protected Configuration

The configuration provided in these keys are stored as encrypted and are only
decrypted inside your Virtual Machine:

* `commandToExecute`: (optional, string) the entrypoint script to execute. Use
  this field instead if your command contains secrets such as passwords.
* `storageAccountName`: (optional, string) the name of storage account. If you
  specify storage credentials, all `fileUris` must be URLs for Azure Blobs.
* `storageAccountKey`: (optional, string) the access key of storage account

json
```json
{
  "commandToExecute": "<command-to-execute>",
  "storageAccountName": "<storage-account-name>",
  "storageAccountKey": "<storage-account-key>"
}
```

PowerShell
```
   $Command2Exec = "<command-to-execute>"
   $StorageAccountName = "<storage-account-name>"
   $StorageAccountKey = "<storage-account-key>"
   $PrivateConf = '{
        "commandToExecute": "'+$Command2Exec+'",
        "storageAccountName": "'+$StorageAccountName+'",
        "storageAccountKey": "'+$StorageAccountKey+'"
    }' 
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



_For PowerShell with variables_
 All PowerShell examples assume you are successfully connected to an Azure subscription. 
   For more information see: http://itproguru.com/expert/2016/04/powershell-working-with-azure-resource-manager-rm-step-by-step-changing-rm-subscriptions/ 
 *You can find values for your variables using various calls such as: 
 *Find VmName : `Get-AzureRmVm -ResourceGroupName $rgName`
 *Find StorageAccount : `Get-AzureRMStorageAccount  -ResourceGroupName $rgName`
 *Find StorageAccountKey:	 
	 +   `$StorageAccount = Get-AzureRmStorageAccount -StorageAccountName $StorageAccountName  -ResourceGroupName $rgName` # Get the Storage Account 
	 +   `$StorageAccountKey = (Get-AzureRmStorageAccountKey -ResourceGroupName $rgName -StorageAccountName $StorageAccountName).Value[0]` # Get the primary Key 
	 +   `$StorageAccountContext = New-AzureStorageContext -StorageAccountKey $StorageAccountKey -StorageAccountName $StorageAccountName` # Get the Context 
	 +   `$UriEndpoint = $StorageAccountContext.BlobEndPoint`  # EndPoint URL
	 + Grab the endpoint URL
	 - `$filename = "hello.sh"`
	 - `$ConfigScriptUri = $StorageAccountContext.BlobEndPoint + "<container>/"+$fileName`
	 - `$ConfigScriptFileName =  $fileName`


```
$Location = "centralus"
$rgName = "<ResourceGroupName>"
$VmName = "<vm-name>"
$ExtensionName = "CustomScriptForLinux"
$Publisher = 'Microsoft.OSTCExtensions'
$Version = "1.5"  # Latest version https://github.com/Azure/custom-script-extension-linux

# Public Configuration ... from above
$ConfigScriptURI="https://gist.github.com/ahmetalpbalkan/b5d4a856fe15464015ae87d5587a4439/raw/466f5c30507c990a4d5a2f5c79f901fa89a80841/hello.sh"
$ConfigScriptFileName = "hello.sh"
$Command2Exec = "sh $ConfigScriptFileName"
$PublicConf = '{
   "fileUris": ["'+$ConfigScriptUri+'"],
   "commandToExecute": "'+$Command2Exec+'"
}'
$StorageAccountName = "<storage-account-name>"
$StorageAccountKey = "<storage-account-key>"
$PrivateConf = '{
     "storageAccountName": "'+$StorageAccountName+'",
     "storageAccountKey": "'+$StorageAccountKey+'"
}' 

# Backtick is the character being used below to continue on next line (uaually left of the 1 on QWERTY keyboard)
# We have all the variables set, let's execute....

Set-AzureRmVMExtension -ResourceGroupName $rgName -VMName $VmName -Location $Location `
  -Name $ExtensionName -Publisher $Publisher `
  -ExtensionType $ExtensionName -TypeHandlerVersion $Version `
  -Settingstring $PublicConf -ProtectedSettingString $PrivateConf

  # Now tell the users where the files are located...
Write-Host "Your Execution Script files are downloaded to: /var/lib/waagent/$Publisher.$ExtensionName-$version.?.?/#/" -ForegroundColor Yellow
Write-Host "    command output is saved to stdout and stderr files in this directory" 
Write-Host "You can find the logs for the extension: "
Write-Host "     /var/log/azure/$Publisher.$ExtensionName/$version/CommandExecution.log " -ForegroundColor Green
Write-Host "     /var/log/azure/$Publisher.$ExtensionName/$version/extension.log" -ForegroundColor Green
  
```

# 3. Troubleshooting

Your files are downloaded to a path like: 
   `/var/lib/waagent/<Publisher>.<ExtensionName>-<version>/#/ScriptName.ext` 
    Example: 
	  `/var/lib/waagent/Microsoft.OSTCExtensions.CustomScriptForLinux-1.5.2.1/download/0/hello.sh` 
the command output is saved to `stdout` and `stderr` files in this directory. Please read
these files to determine output from your script.

You can find the logs for the extension at: 
   `/var/log/azure/<Publisher>.<Extension>/<version>/CommandExecution.log`.
   `/var/log/azure/<Publisher>.<Extension>/<version>/extension.log`.
   Examples:   
    `/var/log/azure/Microsoft.OSTCExtensions.CustomScriptForLinux/1.5.2.1/extension.log`
    `/var/log/azure/Microsoft.OSTCExtensions.CustomScriptForLinux/1.5.2.1/CommandExecution`

_PowerShell Write the locations and examples out to users_
``` 
# Tell the users where the files are located...
Write-Host "Your Execution Script files are downloaded to: /var/lib/waagent/$Publisher.$ExtensionName-$version.?.?/#/" -ForegroundColor Yellow
Write-Host "    command output is saved to stdout and stderr files in this directory" 
Write-host "From command prompt# use cat to display contents of files examples:" -ForegroundColor Green 
Write-host "   sudo cat /var/lib/waagent/Microsoft.OSTCExtensions.CustomScriptForLinux-1.5.2.1/download/0/errout"
Write-host "   sudo cat /var/lib/waagent/Microsoft.OSTCExtensions.CustomScriptForLinux-1.5.2.1/download/0/stdout"
Write-host "   sudo cat /var/lib/waagent/Microsoft.OSTCExtensions.CustomScriptForLinux-1.5.2.1/download/0/hello.sh"

Write-Host "You can find the logs for the extension: "
Write-Host "     /var/log/azure/$Publisher.$ExtensionName/$version/CommandExecution.log " -ForegroundColor Green
Write-Host "     /var/log/azure/$Publisher.$ExtensionName/$version/extension.log" -ForegroundColor Green
Write-host "From command prompt# use cat to display contents of files examples:" -ForegroundColor Green
Write-host "   sudo cat /var/log/azure/Microsoft.OSTCExtensions.CustomScriptForLinux/1.5.2.1/extension.log"
Write-host "   sudo cat /var/log/azure/Microsoft.OSTCExtensions.CustomScriptForLinux/1.5.2.1/CommandExecution.log"

```

   
Please open an issue on this GitHub repository if you encounter problems that
you could not debug with these log files.  

-----

[![Build Status](https://travis-ci.org/Azure/custom-script-extension-linux.svg?branch=master)](https://travis-ci.org/Azure/custom-script-extension-linux)

-----
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
