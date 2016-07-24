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
    [[ "$diff" = *"A /var/lib/azure/custom-script"* ]]
}

@test "handler command: enable - can process empty settings, but fails" {
    mk_container sh -c "fake-waagent install && fake-waagent enable"
    push_settings '' ''

    run start_container
    echo "$output"
    [ "$status" -eq 1 ]
    [[ "$output" == *"invalid configuration: 'commandToExecute' is not specified"* ]]

     # Validate .status file says enable failed
     diff="$(container_diff)"; echo "$diff"
    [[ "$diff" = *"A /var/lib/waagent/Extension/status/0.status"* ]]
    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "$status_file"; [[ "$status_file" = *'Enable failed'* ]]
}

@test "handler command: enable - validates json schema" {
    mk_container sh -c "fake-waagent install && fake-waagent enable"
    push_settings '{"badElement":null, "commandToExecute":"date"}' ''
   
    run start_container
    echo "$output"
    [ "$status" -eq 1 ]
    [[ "$output" == *"json validation error: invalid public settings JSON: badElement"* ]]
}


@test "handler command: enable - captures stdout/stderr into file" {
    mk_container sh -c "fake-waagent install && fake-waagent enable"
    push_settings '
    {
        "commandToExecute": "echo HelloStdout>&1; echo HelloStderr>&2"
    }' ''
    run start_container
    echo "$output"
    [ "$status" -eq 0 ]

    # Validate contents of stdout/stderr files
    stdout="$(container_read_file /var/lib/azure/custom-script/download/0/stdout)"
    echo "stdout=$stdout" && [[ "$stdout" = "HelloStdout" ]]
    stderr="$(container_read_file /var/lib/azure/custom-script/download/0/stderr)"
    echo "stderr=$stderr" && [[ "$stderr" = "HelloStderr" ]]
}

@test "handler command: enable - doesn't process the same sequence number again" {
    mk_container sh -c \
        "fake-waagent install && fake-waagent enable && fake-waagent enable"
    push_settings '{"commandToExecute": "date"}' ''
   
    run start_container
    echo "$output"
    [ "$status" -eq 0 ]
    enable_count="$(echo "$output" | grep -c 'event=enabled')"
    echo "Enable count=$enable_count"
    [ "$enable_count" -eq 1 ]
    [[ "$output" == *"this script configuration is already processed, will not run again"* ]] # not processed again
}

@test "handler command: enable - parses protected settings" {
    mk_container sh -c "fake-waagent install && fake-waagent enable"
    push_settings ''  '{"commandToExecute":"touch /a.txt"}'
    run start_container
    echo "$output"
    [ "$status" -eq 0 ]

    diff="$(container_diff)"; echo "$diff"
    [[ "$diff" == *"A /a.txt"* ]]
}

@test "handler command: enable - downloads files" {
    mk_container sh -c "fake-waagent install && fake-waagent enable"
    # download an external script and run it
    push_settings '{
        "fileUris": [
                "https://gist.github.com/anonymous/8c83af2923ec8dd4a92309594a6c90d7/raw/f26c2cbf68e22d42f703b78f8a4562c5c8e43ba7/script.sh"
        ],
        "commandToExecute":"./script.sh"
        }'
    run start_container
    echo "$output"
    [ "$status" -eq 0 ]

    diff="$(container_diff)"; echo "$diff"
    [[ "$diff" == *"A /var/lib/azure/custom-script/download/0/script.sh"* ]] # file downloaded
    [[ "$diff" == *"A /b.txt"* ]] # created by script.sh
}

@test "handler command: uninstall - deletes the data dir" {
    run in_container sh -c \
        "fake-waagent install && fake-waagent uninstall"
    echo "$output"
    [ "$status" -eq 0 ]

    echo "$output"; [ "$status" -eq 0 ]
    diff="$(container_diff)" && echo "$diff"
    [[ "$diff" != */var/lib/azure/custom-script* ]]
}
