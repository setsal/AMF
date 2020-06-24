.PHONY: \
	build \
	install \
	all \
	vendor \
	lint \
	vet \
	fmt \
	fmtcheck \
	pretest \
	test \
	integration \
	cov \
	clean

SRCS = $(shell git ls-files '*.go')
PKGS =  ./commands ./config ./server ./util
VERSION := 0.0.0
LDFLAGS := -X 'main.version=$(VERSION)' \

all: build test

build: main.go
	go build -ldflags "$(LDFLAGS)" -o go-AMF $<

install: main.go
	go install -ldflags "$(LDFLAGS)"

all: test

clean:
	$(foreach pkg,$(PKGS),go clean $(pkg) || exit;)

