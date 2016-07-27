BINDIR=bin
BIN=custom-script-extension

binary: clean
	if [ -z "$$GOPATH" ]; then \
	  echo "GOPATH is not set"; \
	  exit 1; \
	fi
	GOOS=linux GOARCH=amd64 go build -v \
	  -ldflags "-X main.BuildDate=`date -u --rfc-3339=seconds 2> /dev/null | sed -e 's/ /T/'` \
	  -X main.Version=`<./VERSION` \
	  -X main.GitCommit=`git rev-parse --short HEAD` \
	  -X main.State=`if [ -n "$$(git status --porcelain)" ]; then echo 'dirty'; fi`" \
	  -o $(BINDIR)/$(BIN) . 
clean:
	rm -rf "$(BINDIR)"

.PHONY: clean
