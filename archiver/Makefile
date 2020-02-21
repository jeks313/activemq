PROJ="activemq-archiver"
REPO="github.com/jeks313/activemq-archiver"
VERSION=$(shell git describe --tags --always)
GITHASH=$(shell git rev-parse --short HEAD)
GITBRANCH=$(shell git rev-parse --abbrev-ref HEAD | grep -oE "^([A-Z]+-[0-9]+|master)" | tr '[:upper:]' '[:lower:]')
BUILDDATE=$(shell date '+%Y-%m-%d_%H:%M:%S_%Z')

ifeq ($(GITBRANCH),master)
	BUILDTAG=$(VERSION)
else
	BUILDTAG=$(VERSION)-$(GITBRANCH)
endif

build: *.go
	@mkdir -p build
	@echo $(VERSION) $(GITHASH) $(BUILDDATE)
	go build -o build/$(PROJ) -ldflags "\
		-X $(REPO)/options.Version=$(VERSION) \
	 	-X $(REPO)/options.GitHash=$(GITHASH) \
		-X $(REPO)/options.Build=$(BUILDDATE) \
		"
	cp build/$(PROJ) build/$(PROJ)-$(VERSION)
