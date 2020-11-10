BINDIR=bin
BIN=run-command-handler
BUNDLEDIR=bundle
BUNDLE=run-command-handler.zip

bundle: clean binary
	@mkdir -p $(BUNDLEDIR)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/$(BIN)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/run-command-shim
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/HandlerManifest.json
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/manifest.xml

binary: clean
	if [ -z "$$GOPATH" ]; then \
	  echo "GOPATH is not set"; \
	  exit 1; \
	fi
	go get -d -u -f github.com/Azure/azure-extension-foundation/...
	GOOS=linux GOARCH=amd64 govvv build -v \
	  -ldflags "-X main.Version=`grep -E -m 1 -o  '<Version>(.*)</Version>' misc/manifest.xml | awk -F">" '{print $$2}' | awk -F"<" '{print $$1}'`" \
	  -o $(BINDIR)/$(BIN) ./main 
	cp ./misc/run-command-shim ./$(BINDIR)
clean:
	rm -rf "$(BINDIR)" "$(BUNDLEDIR)"

.PHONY: clean binary
