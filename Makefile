BINDIR := ~/bin
GOPATH := $(shell go env GOPATH)

.PHONY: $(BINDIR)/jgrep
$(BINDIR)/jgrep:
	go build -o $@ ./cmd/jgrep.go

.PHONY: clean
clean:
	rm -f $(BINDIR)/jgrep

test:
	go test -v ./...
