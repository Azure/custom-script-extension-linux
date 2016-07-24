#!/usr/bin/env bats

load test_helper

@test "meta: docker is installed" {
    run docker version
    echo "$output">&2
    [ "$status" -eq 0 ]
}

@test "meta: can build the test container image" {
    run build_docker_image
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "meta: can start the test container" {
    run in_tmp_container fake-waagent
    echo "$output"
    [ "$output" = "Usage: /sbin/fake-waagent <handlerCommand>" ]
    [ "$status" -eq 1 ]
}

@test "meta: can create vm private/public keys" {
    run mk_certs
    echo "$output"
    [ "$status" -eq 0  ]

    thumbprint="$output"
    [ -f "$certs_dir/$thumbprint.prv" ]
    [ -f "$certs_dir/$thumbprint.crt" ]
}

@test "meta: encrypt a protected settings" {
    run mk_certs
    echo "$output"
    [ "$status" -eq 0  ]

    tp="$output"
    run encrypt_settings "$tp" "`seq 0 1000`"
    echo "$output"
    [ "$status" -eq 0 ]
    [ -n "$output" ]
}

@test "meta: can create a .settings json with public/protected config" {
    tp="$mk_certs"
    run mk_settings_json '' '{"commandToExecute":"touch /a.txt"}' "$tp"
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "meta: can create a temporary file" {
    run save_tmp_file "foobar"
    echo "$output"
    [ "$status" -eq 0 ]
    [ -f "$output" ]
    [[ "$(cat "$output")" == "foobar" ]]
}