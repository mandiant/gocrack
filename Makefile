BUILDDATE		:= $(shell date +%s)
BUILDREV		:= $(shell git log --pretty=format:'%h' -n 1)
OS				:= $(shell uname -s| tr '[:upper:]' '[:lower:]')
ARCH			:= amd64
BINDIR			:= builds/$(OS)/$(ARCH)
SERVBUILDTAGS	:= ""
WORKBUILDTAGS	:= ""

ifdef DEBUG
CFLAGS                  += -gcflags '-N -l'
endif

build:
	mkdir -p builds/
	go build -tags "$(WORKBUILDTAGS)" -o $(BINDIR)/gocrack_worker $(CFLAGS) -ldflags \
		"-X github.com/fireeye/gocrack/worker.CompileRev=${BUILDREV} \
		 -X github.com/fireeye/gocrack/worker.CompileTime=${BUILDDATE}" \
		cmd/gocrack_worker/*.go
	go build -tags "$(SERVBUILDTAGS)" -o $(BINDIR)/gocrack_server $(CFLAGS) -ldflags \
		"-X github.com/fireeye/gocrack/server.CompileRev=${BUILDREV} \
		 -X github.com/fireeye/gocrack/server.CompileTime=${BUILDDATE}" \
		cmd/gocrack_server/*.go

static_analysis:
	go vet `go list ./... | egrep -v 'gocat|vendor'`

test:
	go test -cover -v `go list ./... | egrep -v 'gocat|vendor'`

clean:
	rm -rf builds/

all: static_analysis test build