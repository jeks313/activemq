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

.PHONY: build
build: archiver consumer producer

.PHONY: get
get:
	@mkdir -p build
	@go mod download

.PHONY: archiver
archiver: get
	echo $(VERSION) $(GITHASH) $(BUILDDATE)
	go build -o build/activemq-archiver -ldflags "\
		-X $(REPO)/pkg/options.Version=$(VERSION) \
	 	-X $(REPO)/pkg/options.GitHash=$(GITHASH) \
		-X $(REPO)/pkg/options.Build=$(BUILDDATE) \
		" ./cmd/archiver/.
	cp build/activemq-archiver build/activemq-archiver-$(VERSION)

.PHONY: producer
producer: get
	echo $(VERSION) $(GITHASH) $(BUILDDATE)
	go build -o build/producer -ldflags "\
		-X $(REPO)/pkg/options.Version=$(VERSION) \
	 	-X $(REPO)/pkg/options.GitHash=$(GITHASH) \
		-X $(REPO)/pkg/options.Build=$(BUILDDATE) \
		" ./cmd/producer/.
	cp build/producer build/producer-$(VERSION)

.PHONY: consumer
consumer: get
	echo $(VERSION) $(GITHASH) $(BUILDDATE)
	go build -o build/consumer -ldflags "\
		-X $(REPO)/pkg/options.Version=$(VERSION) \
	 	-X $(REPO)/pkg/options.GitHash=$(GITHASH) \
		-X $(REPO)/pkg/options.Build=$(BUILDDATE) \
		" ./cmd/consumer/.
	cp build/consumer build/consumer-$(VERSION)
