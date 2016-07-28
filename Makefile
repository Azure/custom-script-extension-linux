BINDIR=bin
BIN=custom-script-extension
BUNDLEDIR=bundle
BUNDLE=custom-script-extension.zip

binary: clean
	if [ -z "$$GOPATH" ]; then \
	  echo "GOPATH is not set"; \
	  exit 1; \
	fi
	GOOS=linux GOARCH=amd64 go build -v \
	  -ldflags "-X main.BuildDate=`date -u --rfc-3339=seconds 2> /dev/null | sed -e 's/ /T/'` \
	  -X main.Version=`<./VERSION` \
	  -X main.GitCommit=`git rev-parse --short HEAD` \
	  -X main.State=`if [[ -n $$(git status --porcelain) ]]; then echo 'dirty'; fi`" \
	  -o $(BINDIR)/$(BIN) . 
	cp ./misc/custom-script-shim ./$(BINDIR)
bundle: clean binary
	@mkdir -p $(BUNDLEDIR)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/$(BIN)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/custom-script-shim
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./HandlerManifest.json
clean:
	rm -rf "$(BINDIR)" "$(BUNDLEDIR)"

.PHONY: clean binary
