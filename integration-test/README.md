# Integration Tests

This directory contains instructions and test files on how to test the extension
binary in an environment that looks like the waagent (Azure Linux Guest Agent)
has just unpackaged it.


## Requirements

- Install Docker Engine on Linux or Docker for Mac.

## Testing

Change working directory to the repository root.

Build the extension binary with `make bundle`.

Build the test base image:

    docker build -f test.Dockerfile -t custom-script .

(The image needs to be built again every time you have a new binary.)

Create an start a container:

    docker run -i -t --rm custom-script

Now you are in the container.

Now just run `waagent <command>` to have the fake agent invoke the extension
handler:

    # fake-waagent enable
    (reads HandlerManifest.json for "enableCommand")
    (calls "enableCommand" and prints its output)
    (now you inspect the filesystem to validate the agent)

Once you exit the bash prompt, the container will be deleted (`--rm`).