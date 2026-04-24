BINDIR=bin
BIN=custom-script-extension
BIN_ARM64=custom-script-extension-arm64
BUNDLEDIR=bundle
BUNDLE=custom-script-extension.zip
GOPATH ?= $(shell go env GOPATH 2>/dev/null)
VERSION ?=

bundle: clean binary
	@mkdir -p $(BUNDLEDIR)
	zip -r ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/HandlerManifest.json

binary: clean
	if [ -z "$(VERSION)" ]; then \
	  echo "VERSION is required. Usage: make VERSION=X.Y.Z"; \
	  exit 1; \
	fi
	if [ -z "$(GOPATH)" ] || [ ! -d "$(GOPATH)" ]; then \
	  echo "GOPATH from 'go env GOPATH' is not set or does not exist: $(GOPATH)"; \
	  exit 1; \
	fi
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v \
	  -tags "netgo osusergo" \
	  -ldflags "-X main.Version=$(VERSION)" \
	  -o $(BINDIR)/$(BIN) ./main
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v \
	  -tags "netgo osusergo" \
	  -ldflags "-X main.Version=$(VERSION)" \
	  -o $(BINDIR)/$(BIN_ARM64) ./main 
	cp ./misc/custom-script-shim ./$(BINDIR)

clean:
	rm -rf "$(BINDIR)" "$(BUNDLEDIR)"

.PHONY: clean binary

