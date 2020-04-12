
# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH))")

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BINARY=doig
VERSION=0.1
BUILD=`git rev-parse HEAD`
PLATFORMS=darwin linux windows
ARCHITECTURES=386 amd64

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

default: build

all: clean build_all install

build:
	go build ${LDFLAGS} -o ${BINARY}

release: build
	mkdir ${BINARY}-${VERSION}
	mv ${BINARY} ${BINARY}-${VERSION}/
	cp Dockerfile.template ${BINARY}-${VERSION}/
	cp -R tools ${BINARY}-${VERSION}/
	tar czvf  ${BINARY}-${VERSION}.tgz ${BINARY}-${VERSION}
	rm -rf ${BINARY}-${VERSION}

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v -o $(BINARY)-$(GOOS)-$(GOARCH))))

release_all: build_all
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); \
	mkdir $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION); \
	mv $(BINARY)-$(GOOS)-$(GOARCH) $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION)/; \
	cp Dockerfile.template $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION)/; \
	cp -R tools $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION)/; \
	tar czvf $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION).tgz \
	$(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION) Dockerfile.template tools; \
	rm -rf $(BINARY)-$(GOOS)-$(GOARCH)-$(VERSION))))

install:
	go install ${LDFLAGS}

# Remove only what we've created
clean:
	find ${ROOT_DIR} -name '${BINARY}[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -delete

.PHONY: check clean install build_all all