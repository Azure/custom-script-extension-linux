# Integration Tests

This directory contains instructions and test files on how to test the extension
binary in an environment that looks like an Azure Linux Virtual Machine.

## Requirements

- Install _Docker Engine on Linux_ or _Docker for Mac_.
- Install [Bats](https://github.com/sstephenson/bats).

## Testing Strategy

The integation tests use `bats` to run the scenarios as bash scripts.

To test the extension handler functionality, we simply:

- build a Docker image using test.Dockerfile
    - copy some files to make it look like a `/var/lib/waagent` dir
    - copy extension binary into the container
- remove the `test` container if it exists
- create a Docker container (name: `test`) from image
    - specify which handler subcommand will be invoked (e.g. `fake-waagent
      install`)
- push .settings file and certificate/private key (.crt, .prv)
    - do other things on the container that we need to craft the environment
- start the container
- collect the output from the command execution
- validate using the following:
    - check status code
    - validate output of the command
    - `docker diff test` to validate file changes in the container
    - copy files out of container and validate their contents

## Running Tests

To run the integration tests, run the following commands from the repository
root:

```
make binary
bats integration-test/test
```
