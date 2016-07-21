# Test files

This read-only directory is what is normally under the `/var/lib/waagent`
directory on an Azure Linux VM running the `waagent`.

It only has the files required for extension handler to parse its
configuration and run.

The extension handler binary should be placed at `./Extension/bin/`
directory.

## Files

```
.
├── {THUMBPRINT}.crt            <-- certificate comes from the wire server
├── {THUMBPRINT}.prv            <-- private key comes from the wire server
└── Extension/                  <-- the 'HandlerManifest.json' should go here
    ├── HandlerEnvironment.json <-- handler binary reads this
    ├── bin/                    <-- the 'handler binary' should go here
    ├── config/
    │   └── 0.settings          <-- handler binary reads this
    └── status/                 <-- handler binary should write here
```