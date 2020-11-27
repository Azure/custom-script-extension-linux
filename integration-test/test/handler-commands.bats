#!/usr/bin/env bats

load test_helper

setup(){
    build_docker_image
}

teardown(){
    rm -rf "$certs_dir"
}

@test "handler command: install - creates the data dir" {
    run in_container fake-waagent install
    echo "$output"
    [ "$status" -eq 0 ]
    [[ "$output" = *event=installed* ]]

    diff="$(container_diff)"
    echo "$diff"
    [[ "$diff" = *"A /var/lib/waagent/run-command-handler"* ]]
}

@test "handler command: enable - can process empty settings, but fails" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '' ''

    run start_container
    echo "$output"
    [[ "$output" == *"invalid configuration: Either 'source.script' or 'source.scriptUri' has to be specified"* ]]

     # Validate .status file says enable failed
     diff="$(container_diff)"; echo "$diff"
    [[ "$diff" = *"A /var/lib/waagent/Extension/status/0.status"* ]]
    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "$status_file"
    [[ "$status_file" = *'Execution failed'* ]]
    [[ "$status_file" = *'Running'* ]]
    [[ "$status_file" = *"invalid configuration: Either 'source.script' or 'source.scriptUri'"* ]]
}

@test "handler command: enable - validates json schema" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '{"badElement":null, "source": {"script":"date"}}' ''
   
    run start_container
    echo "$output"
    [[ "$output" == *"json validation error: invalid public settings JSON: badElement"* ]]
}

@test "handler command: enable - captures stdout/stderr into file and .status" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '
    {
        "source": {
            "script": "echo HelloStdout>&1; echo HelloStderr>&2"
        }
    }' ''
    run start_container
    echo "$output"

    # Validate contents of stdout/stderr files
    stdout="$(container_read_file /var/lib/waagent/run-command-handler/download/0/stdout)"
    echo "stdout=$stdout" && [[ "$stdout" = "HelloStdout" ]]
    stderr="$(container_read_file /var/lib/waagent/run-command-handler/download/0/stderr)"
    echo "stderr=$stderr" && [[ "$stderr" = "HelloStderr" ]]

    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "status_file=$status_file"
    [[ "$status_file" = *'Execution succeeded'* ]]
    [[ "$status_file" = *'HelloStdout'* ]]
    [[ "$status_file" = *'HelloStderr'* ]]
}

@test "handler command (multiconfig): enable - captures stdout/stderr into file and .status" {
    # export ConfigExtensionName=extname && export ConfigSequenceNumber=5 will be read from the extension to determine the settings file name
    mk_container sh -c "export ConfigExtensionName=extname && export ConfigSequenceNumber=5 && fake-waagent install && fake-waagent enable && wait-for-enable "
    push_settings '
    {
        "source": {
            "script": "echo HelloStdout>&1; echo HelloStderr>&2"
        }
    }' '' 'extname.5.settings'
    run start_container
    echo "$output"

    # Validate contents of stdout/stderr files
    stdout="$(container_read_file /var/lib/waagent/run-command-handler/download/extname/5/stdout)"
    echo "stdout=$stdout" && [[ "$stdout" = "HelloStdout" ]]
    stderr="$(container_read_file /var/lib/waagent/run-command-handler/download/extname/5/stderr)"
    echo "stderr=$stderr" && [[ "$stderr" = "HelloStderr" ]]

    config_file="$(container_read_file /var/lib/waagent/Extension/config/extname.5.settings)"
    echo "config_file=$config_file"
    [[ "$config_file" = *'echo HelloStdout>&1; echo HelloStderr>&2'* ]]

    status_file="$(container_read_file /var/lib/waagent/Extension/status/extname.5.status)"
    echo "status_file=$status_file"
    [[ "$status_file" = *'Execution succeeded'* ]]
    [[ "$status_file" = *'HelloStdout'* ]]
    [[ "$status_file" = *'HelloStderr'* ]]

    mrseq_file="$(container_read_file /var/lib/waagent/extname.mrseq)"
    echo "mrseq_file=$mrseq_file"
    [[ "$mrseq_file" = '5' ]]
}

@test "handler command: enable - captures stdout/stderr into .status on error" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '
    {
        "source": {
            "script": "ls /does-not-exist"
        }
    }' ''
    run start_container
    echo "$output"

    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "status_file=$status_file"
    [[ "$status_file" = *'Execution succeeded'* ]]
    [[ "$status_file" = *'ls: cannot access'* ]]
}

@test "handler command: enable - doesn't process the same sequence number again" {
    mk_container sh -c \
        "fake-waagent install && fake-waagent enable && wait-for-enable && fake-waagent enable && wait-for-enable"
    push_settings '{"source": {"script":"date"}}' ''
   
    run start_container
    echo "$output"
    enable_count="$(echo "$output" | grep -c 'event=enabled')"
    echo "Enable count=$enable_count"
    [ "$enable_count" -eq 1 ]
    [[ "$output" == *"the script configuration has already been processed, will not run again"* ]] # not processed again
}

@test "handler command: enable - parses public settings" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '{"source": {"script":"touch /a.txt"}}' ''
    run start_container
    echo "$output"

    diff="$(container_diff)"; echo "$diff"
    [[ "$diff" == *"A /a.txt"* ]]
}

@test "handler command: asyncExecution is true" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '
    {
        "asyncExecution": true,
        "source": {"script": "echo Hello"}
    }' ''
    run start_container
    echo "$output"

    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "status_file=$status_file"
    [[ "$status_file" = *'Execution succeeded'* ]]
    [[ "$status_file" = *'Hello'* ]]
}

@test "handler command: enable - downloads files" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    # download an external script and run it
    push_settings '{
            "source": {
                "scriptUri": "https://github.com/koralski/run-command-extension-linux/raw/master/integration-test/testdata/script.sh"
            }
        }'
    run start_container
    echo "$output"

    diff="$(container_diff)"; echo "$diff"
    [[ "$diff" == *"A /var/lib/waagent/run-command-handler/download/0/script.sh"* ]] # file downloaded
    [[ "$diff" == *"A /b.txt"* ]] # created by script.sh
}

# @test "handler command: enable - download files from storage account" {
#     if [[ -z  "$AZURE_STORAGE_ACCOUNT" ]] || [[ -z  "$AZURE_STORAGE_ACCESS_KEY" ]]; then
#         skip "AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY not specified"
#     fi

#     # make random file
#     tmp="$(mktemp)"
#     dd if=/dev/urandom of="$tmp" bs=1k count=512

#     # upload files to a storage container
#     cnt="testcontainer"
#     blob1="blob1-$RANDOM"
#     blob2="blob2 with spaces-$RANDOM"
#     azure storage container show  "$cnt" 1>/dev/null ||
#         azure storage container create "$cnt" 1>/dev/null && echo "Azure Storage container created">&2
#     azure storage blob upload -f "$tmp" "$cnt" "$blob1" 1>/dev/null  # upload blob1
#     azure storage blob upload -f "$tmp" "$cnt" "$blob2" 1>/dev/null # upload blob2

#     blob1_url="http://$AZURE_STORAGE_ACCOUNT.blob.core.windows.net/$cnt/$blob1" # over http
#     blob2_url="https://$AZURE_STORAGE_ACCOUNT.blob.core.windows.net/$cnt/$blob2" # over https

#     mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable" # add sleep for enable to finish in the background
#     # download an external script and run it
#     push_settings '{
#             "source": {
#                 "scriptUri": "$blob1_url"
#             }
#         }' ''
#     run start_container
#     echo "$output"
#     [ "$status" -eq 0 ]
#     [[ "$output" == *'file=0 event="download complete"'* ]]
#     [[ "$output" == *'file=1 event="download complete"'* ]]

#     diff="$(container_diff)"; echo "$diff"
#     [[ "$diff" == *"A /var/lib/waagent/run-command-handler/download/0/$blob1"* ]] # file downloaded
#     [[ "$diff" == *"A /var/lib/waagent/run-command-handler/download/0/$blob2"* ]] # file downloaded

#     # compare checksum
#     existing=$(md5 -q "$tmp")
#     echo "Local file checksum: $existing"
#     got=$(container_read_file "/var/lib/waagent/run-command-handler/download/0/$blob1" | md5 -q)
#     echo "Downloaded file checksum: $got"
#     [[ "$existing" == "$got" ]]
# }


@test "handler command: enable - forking into background does not overwrite existing status" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable && fake-waagent enable && wait-for-enable"
    push_settings '{"source": {"script": "date"}}' ''
    run start_container
    echo "$output"

    # validate .status file still reads "Enable succeeded"
    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "$status_file"
    [[ "$status_file" = *'Execution succeeded'* ]]
}

@test "handler command: uninstall - deletes the data dir" {
    run in_container sh -c \
        "fake-waagent install && fake-waagent uninstall"
    echo "$output"
    [ "$status" -eq 0 ]

    diff="$(container_diff)" && echo "$diff"
    [[ "$diff" != */var/lib/waagent/run-command* ]]
}
