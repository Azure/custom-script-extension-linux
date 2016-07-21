FROM debian:jessie

RUN apt-get -qqy update && \
	apt-get -qqy install jq openssl ca-certificates && \
        apt-get -qqy clean && \
        rm -rf /var/lib/apt/lists/*

# Create the directories that need to be present
RUN mkdir -p /var/lib/waagent && \
        mkdir -p /var/log/azure/Extension/VE.RS.ION 


# Copy the test environment
WORKDIR /var/lib/waagent
COPY integration-test/env/ .
RUN ln -s /var/lib/waagent/fake-waagent /sbin/fake-waagent

# Copy the handler files
COPY HandlerManifest.json ./Extension/
COPY bin/custom-script-extension ./Extension/bin/
