FROM debian:jessie

RUN apt-get -qqy update && \
	apt-get -qqy install jq openssl ca-certificates && \
        apt-get -qqy clean && \
        rm -rf /var/lib/apt/lists/*

# Create the directories and files that need to be present
RUN mkdir -p /var/lib/waagent && \
        mkdir -p /var/lib/waagent/Extension/config && \
        mkdir -p /var/lib/waagent/Extension/status && \
        mkdir -p /var/log/azure/Extension/VE.RS.ION

# Copy the test environment
WORKDIR /var/lib/waagent
COPY integration-test/env/ .
RUN ln -s /var/lib/waagent/fake-waagent /sbin/fake-waagent && \
        ln -s /var/lib/waagent/wait-for-enable /sbin/wait-for-enable

# Copy the handler files
COPY misc/HandlerManifest.json ./Extension/
COPY misc/run-command-shim ./Extension/bin/
COPY bin/run-command-handler ./Extension/bin/
