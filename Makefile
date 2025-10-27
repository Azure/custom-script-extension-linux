BINDIR=bin
BIN=custom-script-extension
BIN_ARM64=custom-script-extension-arm64
BUNDLEDIR=bundle
BUNDLE=custom-script-extension.zip

bundle: clean binary
	@mkdir -p $(BUNDLEDIR)
	zip -r ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/HandlerManifest.json
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/manifest.xml

binary: clean
	if [ -z "$$GOPATH" ]; then \
	  echo "GOPATH is not set"; \
	  exit 1; \
	fi
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v \
	  -tags "netgo osusergo" \
	  -ldflags "-X main.Version=`grep -E -m 1 -o  '<Version>(.*)</Version>' misc/manifest.xml | awk -F">" '{print $$2}' | awk -F"<" '{print $$1}'`" \
	  -o $(BINDIR)/$(BIN) ./main
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v \
	  -tags "netgo osusergo" \
	  -ldflags "-X main.Version=`grep -E -m 1 -o  '<Version>(.*)</Version>' misc/manifest.xml | awk -F">" '{print $$2}' | awk -F"<" '{print $$1}'`" \
	  -o $(BINDIR)/$(BIN_ARM64) ./main 
	cp ./misc/custom-script-shim ./$(BINDIR)
clean:
	rm -rf "$(BINDIR)" "$(BUNDLEDIR)"
.PHONY: clean binary

