# Azure Custom Script Virtual Machine Extension (2.0) 
[![Build Status](https://travis-ci.org/Azure/custom-script-extension-linux.svg?branch=master)](https://travis-ci.org/Azure/custom-script-extension-linux)

Image | Status |
------|--------|
Canonical UbuntuServer 14.04.5-LTS|![badge](https://dcrbadges.blob.core.windows.net/scenarios/Canonical_UbuntuServer_14.04.5-LTS__ext--custom--script.svg)
Canonical UbuntuServer 16.04-LTS|![badge](https://dcrbadges.blob.core.windows.net/scenarios/Canonical_UbuntuServer_16.04-LTS__ext--custom--script.svg)
Canonical UbuntuServer 18.04-LTS|![badge](https://dcrbadges.blob.core.windows.net/scenarios/Canonical_UbuntuServer_18.04-LTS__ext--custom--script.svg)
Credativ Debian 8|![badge](https://dcrbadges.blob.core.windows.net/scenarios/Credativ_Debian_8__ext--custom--script.svg)
Credativ Debian 9|![badge](https://dcrbadges.blob.core.windows.net/scenarios/Credativ_Debian_9__ext--custom--script.svg)
OpenLogic CentOS 6.9|![badge](https://dcrbadges.blob.core.windows.net/scenarios/OpenLogic_CentOS_6.9__ext--custom--script.svg)
OpenLogic CentOS 7.4|![badge](https://dcrbadges.blob.core.windows.net/scenarios/OpenLogic_CentOS_7.4__ext--custom--script.svg)
RedHat RHEL 6.9|![badge](https://dcrbadges.blob.core.windows.net/scenarios/RedHat_RHEL_6.9__ext--custom--script.svg)
RedHat RHEL 7-RAW|![badge](https://dcrbadges.blob.core.windows.net/scenarios/RedHat_RHEL_7-RAW__ext--custom--script.svg)
SUSE SLES 12-SP2|![badge](https://dcrbadges.blob.core.windows.net/scenarios/SUSE_SLES_12-SP2__ext--custom--script.svg)

This documentation is current for version 2.1.3 and above.

CustomScript extensions runs scripts on VMs.  These scripts can be
used to bootstrap/install software, run administrative tasks, or run
automation tasks. CustomScript can run an inline script you specify or
download a script file from the Internet or Azure Storage.

You can add CustomScript extension to your VM using:

- Azure CLI (python based / Cloud Shell)
- Azure XPlat CLI (node based)
- Azure PowerShell
- Azure Resource Manager (ARM) Templates
- Azure Virtual Machines REST API

:information_source: Please read the [Using the Azure Custom Script Extension with Linux
Virtual Machines][doc] page for detailed usage instructions.

[doc]: https://docs.microsoft.com/en-us/azure/virtual-machines/extensions/custom-script-linux

# Custom Script Extension Reference Guide

## 1. Extension Configuration

You can specify files to download and commands to execute in the
configuration section of the extensions. If your `commandToExecute`,
`script`, or `fileUris` contain secrets, please use protected settings
instead of public settings.

The specified command will be executed only once. If you change
anything in the extension configuration and deploy it again, the
command will be executed again.

> If you would like to execute the same command again you must updated
> the configuration otherwise the command will not re-executed.  The
> easiest way to do this is with the timestamp setting. Simply
> increment the timestamp value to re-execute the command.

### 1.1. Public Settings

Schema for the public configuration file looks like this:

* `commandToExecute`: (**required** if script not set, string) the entry point script to execute
* `script`: (**required** if commandToExecute not set, string) a base64 encoded (and optionally gzip'ed) script executed by /bin/sh.
* `skipDos2Unix`: (optional, boolean) skip dos2unix conversion of script-based file URLs or script.
* `fileUris`: (optional, string array) the URLs for file(s) to be downloaded.
* `timestamp` (optional, 32-bit integer) use this field only to trigger a re-run of the
  script by changing value of this field.  Any integer value is acceptable; it must only be different than the previous value.
 
```json
{
  "fileUris": ["<url>"],
  "commandToExecute": "<command-to-execute>"
}
```

> Examples:
>
> ```json
> {
>   "fileUris": ["https://gist.github.com/ahmetalpbalkan/b5d4a856fe15464015ae87d5587a4439/raw/466f5c30507c990a4d5a2f5c79f901fa89a80841/hello.sh"],
>   "commandToExecute": "./hello.sh"
> }
> ```
> 
> ```json
> {
>   "commandToExecute": "apt-get -y update && apt-get install -y apache2"
> }
> ```


### 1.2. Protected Settings

The configuration provided in these keys are stored encrypted and are only
decrypted inside your VM.

* `commandToExecute`: (optional, string) the entry point script to execute. Use
  this field instead if your command contains secrets such as passwords.
* `script`: (optional, string) a base64 encoded (and optionally gzip'ed) script executed by /bin/sh.
* `fileUris`: (optional, string array) the URLs for file(s) to be downloaded.
* `storageAccountName`: (optional, string) the name of storage account. If you
  specify storage credentials, all `fileUris` must be URLs for Azure Blobs.
* `storageAccountKey`: (optional, string) the access key of storage account
* `managedIdentity`: (optional, json object) the [managed identity](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview) for downloading file(s)
  * `clientId`: (optional, string) the client id of the managed identity
  * `objectId`: (optional, string) the object id of the managed identity


```json
{
  "commandToExecute": "<command-to-execute>",
  "storageAccountName": "<storage-account-name>",
  "storageAccountKey": "<storage-account-key>",
  "managedIdentity": "<managed-identity>"
}
```

### 1.3. Public vs. Protected Settings

Public settings are sent in clear text to the VM where the script will
be executed.  Protected settings are encrypted using a key known only
to the Azure and the VM. The settings are saved to the VM as they were
sent, i.e. if the settings were encrypted they are saved encrypted on
the VM. The certificate used to decrypt the encrypted values is stored
on the VM, and used to decrypt settings (if necessary) at runtime.

The following values can be set in either public or protected
settings.

* `commandToExecute`
* `script`
* `fileUris`

The extension will reject any configuration where the above values are
set in both public and protected settings.

The following values can only be set in **public** settings.

* `skipDos2Unix`
* `timestamp`

The follow values can only by set in **protected** settings.

* `storageAccountName`
* `storageAccountKey`
* `managedIdentity`

### 1.3 skipDos2Unix

The default value is false, which means dos2unix conversion **is**
executed.

The previous version of CustomScript,
Microsoft.OSTCExtensions.CustomScriptForLinux, would automatically
convert DOS files to UNIX files by translating `\r\n` to `\n`.  This
translation still exists, and is on by default.  This conversion is
applied to all files downloaded from fileUris or the script setting
based on any of the following criteria.

* If the extension is one of `.sh`, `.txt`, `.py`, or `.pl` it will be
  converted.  The script setting will always match this criteria
  because it is assumed to be a script executed with /bin/sh, and is
  saved as script.sh on the VM.
* If the file starts with `#!`.

The dos2unix conversion can be skipped by setting the skipDos2Unix to
true.

```json
{
  "fileUris": ["<url>"],
  "commandToExecute": "<command-to-execute>",
  "skipDos2Unix": true
}
```

### 1.4 script

CustomScript supports execution of a user-defined script.  The script
settings to combine commandToExecute and fileUris into a single
setting.  Instead of the having to setup a file for download from
Azure storage or GitHub gist, you can simply encode the script as a
setting.  Script can be used to replaced commandToExecute and
fileUris.

The script **must** be base64 encoded.  The script can **optionally**
be gzip'ed.  The script setting can be used in public or protected
settings. The maximum size of the script parameter's data is 256
KB. If the script exceeds this size it will not be executed.

For example, given the following script saved to the file /script.sh/.

```sh
#!/bin/sh
echo "Updating packages ..."
apt update
apt upgrade -y
```

The correct CustomScript script setting would be constructed by taking
the output of the following command.

```sh
cat script.sh | base64 -w0
```

```json
{
  "script": "IyEvYmluL3NoCmVjaG8gIlVwZGF0aW5nIHBhY2thZ2VzIC4uLiIKYXB0IHVwZGF0ZQphcHQgdXBncmFkZSAteQo="
}
```

The script can optionally be gzip'ed to further reduce size (in most
cases).  (CustomScript auto-detects the use of gzip compression.)

```sh
cat script | gzip -9 | base64 -w 0
```

CustomScript uses the following algorithm to execute a script.

 1. assert the length of the script's value does not exceed 256 KB.
 1. base64 decode the script's value
 1. _attempt_ to gunzip the base64 decoded value
 1. write the decoded (and optionally decompressed) value to disk (/var/lib/waagent/custom-script/#/script.sh)
 1. execute the script using _/bin/sh -c /var/lib/waagent/custom-script/#/script.sh.

### 1.5 managedIdentity
> [NOTE]
> This property must be specified in protected settings only. 

CustomScript (version 2.1 onwards) supports [managed identity](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview) for downloading file(s) from URLs provided in the "fileUris" setting. It allows CustomScript to access private Azure Storage blobs or containers without the user having to pass secrets like SAS tokens or storage account keys.

To use this feature, the user must add a [system-assigned](https://docs.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=dotnet#adding-a-system-assigned-identity) or [user-assigned](https://docs.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=dotnet#adding-a-user-assigned-identity) identity to the VM or VMSS where CustomScript is expected to run, and [grant the managed identity access to the Azure Storage container or blob](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/tutorial-vm-windows-access-storage#grant-access).

To use the system-assigned identity on the target VM/VMSS, set "managedidentity" field to an empty json object. 

> Example:
>
> ```json
> {
>   "fileUris": ["https://mystorage.blob.core.windows.net/privatecontainer/script1.sh"],
>   "commandToExecute": "sh script1.sh",
>   "managedIdentity" : {}
> }
> ```

To use the user-assigned identity on the target VM/VMSS, configure "managedidentity" field with the client id or the object id of the managed identity.

> Examples:
>
> ```json
> {
>   "fileUris": ["https://mystorage.blob.core.windows.net/privatecontainer/script1.sh"],
>   "commandToExecute": "sh script1.sh",
>   "managedIdentity" : { "clientId": "31b403aa-c364-4240-a7ff-d85fb6cd7232" }
> }
> ```
> ```json
> {
>   "fileUris": ["https://mystorage.blob.core.windows.net/privatecontainer/script1.sh"],
>   "commandToExecute": "sh script1.sh",
>   "managedIdentity" : { "objectId": "12dd289c-0583-46e5-b9b4-115d5c19ef4b" }
> }
> ```

> [NOTE]
> managedIdentity property **must not** be used in conjunction with storageAccountName or storageAccountKey properties

# 2. Deployment to a Virtual Machine

For **ARM templates**, see [this documentation][doc] to create an extension
resource in your template.

[doc]: https://azure.microsoft.com/documentation/articles/virtual-machines-linux-extensions-customscript/

For **Azure CLI**, create a `public.json` (and optionally `protected.json`) and run:

    $ az vm extension set --resource-group <resource-group> --vm-name <vm-name> \
        --name CustomScript --publisher Microsoft.Azure.Extensions --version 2.0 \
        --settings ./public.json \
        --protected-settings ./protected.json

For **Azure XPlat CLI**, create a `public.json` (and optionally `protected.json`) and run:

    $ azure vm extension set <resource-group> <vm-name> \
	    CustomScript Microsoft.Azure.Extensions 2.0 \
	    --auto-upgrade-minor-version \
	    --public-config-path public.json \
	    --private-config-path protected.json

# 3. Troubleshooting

Your files are downloaded to a path like: `/var/lib/waagent/custom-script/download/0/` and
the command output is saved to `stdout` and `stderr` files in this directory. Please read
these files to find out output from your script.

You can find the logs for the extension at `/var/log/azure/custom-script/handler.log`.

Please open an issue on this GitHub repository if you encounter problems that
you could not debug with these log files.

# 4. Build Badges

The Custom Script extension is tested across a variety of distributions.
There are tests cases to exercise all of the settings supported by
Custom Script.  This includes the following:

 1. public and protected commandToExecute 
 1. public and protected script
 1. public and protected script that has been gzip'ed
 1. public and protected fileUris
 1. invalid script
 1. scripts that are too big are rejected
 
 Tests are executed with the version that ships with the Linux distro being tested.

-----
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
