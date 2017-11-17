# Test helpers for BATS tests

IMAGE=run-command
DOCKERFILE=test.Dockerfile
TEST_CONTAINER=test

certs_dir="$BATS_TEST_DIRNAME/certs"

build_docker_image() {
    echo "Building test image..."
    docker build -q -f $DOCKERFILE -t $IMAGE . 1>&2
}

in_tmp_container() {
    docker run --rm $IMAGE "$@"
}

rm_container() {
    docker rm -f $TEST_CONTAINER &>/dev/null && \
        echo "Deleted test container." || true
}

mk_container() {
    rm_container && echo "Creating test container with commands: $@">&2 && \
        docker create --name=$TEST_CONTAINER $IMAGE "$@" 1>/dev/null
}

in_container() {
    set -e
    rm_container
    mk_container "$@"
    echo "Starting test container...">&2
    start_container
}

start_container() {
    docker start --attach $TEST_CONTAINER
}

container_diff() {
    docker diff $TEST_CONTAINER
}

container_read_file() { # reads the file at container path $1
    set -eo pipefail
    docker cp $TEST_CONTAINER:"$1" - | tar x --to-stdout
} 

mk_certs() { # creates certs/{THUMBPRINT}.(crt|key) files under ./certs/ and prints THUMBPRINT
    set -eo pipefail
    mkdir -p "$certs_dir" && cd "$certs_dir" && rm -f "$certs_dir/*"
    openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -batch &>/dev/null
    thumbprint=$(openssl x509 -in cert.pem -fingerprint -noout| sed 's/.*=//g' | sed 's/://g')
    mv cert.pem $thumbprint.crt && \
        mv key.pem $thumbprint.prv
    echo $thumbprint
}

push_certs() { # pushes certs/{$1}.(crt|key) files to container
    set -e
    docker cp "$certs_dir/$1.crt" $TEST_CONTAINER:/var/lib/waagent/
    docker cp "$certs_dir/$1.prv" $TEST_CONTAINER:/var/lib/waagent/
    echo "Pushed certs to container.">&2
}

encrypt_settings(){ # encrypts the message ($2) with the key with given cert thumbprint ($1)
    set -eo pipefail
    tp="$1"; msg="$2"
    echo "$(openssl smime -inkey "$certs_dir/$tp.prv" -encrypt -outform DER "$certs_dir/$tp.crt" < <(echo "$msg") | base64 -w0)" 
}

mk_settings_json() { # turns json public settings ($1) and ($2) into a json encrypted with "$3".(crt|prv)
    set -e
    pub="$1"
    prot="$2"
    cert_tp="$3"
    if [ -z "$pub" ]; then pub="null"; fi    
    if [ -n "$prot" ]; then
        prot="\"$(encrypt_settings "$cert_tp" "$prot")\""
    else
        cert_tp="null"
        prot="null"
    fi
    
	cat <<-EOF
	{
        "runtimeSettings": [
            {
                "handlerSettings": {
                    "protectedSettingsCertThumbprint": "$cert_tp",
                    "publicSettings": $pub,
                    "protectedSettings": $prot
                }
            }
        ]
	}
EOF
}

push_settings() { # creates and copies 0.settings file with given public settings ($1) and ($2) values.
    set -e

    if [ -n "$2" ]; then
        cert_tp="$(mk_certs)"
        push_certs "$cert_tp"
    fi

    cfg_file="$(save_tmp_file "$(mk_settings_json "$1" "$2" "$cert_tp") ")"
    echo ".settings: $(cat "$cfg_file")" >&2
    copy_config "$cfg_file"
    echo ".settings file pushed to container." >&2
}

save_tmp_file(){ # saves $1 into a temporary file and returns its path
    fp="$(mktemp)"
    touch "$fp"
    cat <<< "$1" > "$fp"
    echo "$fp"
}

copy_config() { # places specified settings file ($1) into container as 0.settings
    set -e
    echo "Copying $1 to container as 0.settings." >&2
    docker cp "$1" "$TEST_CONTAINER:/var/lib/waagent/Extension/config/0.settings"
    echo "Copied settings into container.">&2
}