# Integration Test Environment

This directory holds a skeleton of what we copy to `/var/lib/waagent` in the
integration testing Docker image.

```
.
├── {THUMBPRINT}.crt            <-- tests generate and push this certificate
├── {THUMBPRINT}.prv            <-- tests generate and push this private key
└── Extension/                  
    ├── HandlerManifest.json    <-- docker image build pushes it here
    ├── HandlerEnvironment.json <-- the extension reads this
    ├── bin/                    <-- docker image build pushes the extension binary here
    ├── config/                 <-- tests push [{extName}.]{seqNo}.settings file here
    └── status/                 <-- extension should write here [{extName}.]{seqNo}.status
```