SHELL:=/bin/bash

build: build-tweet

build-tweet:
	$(eval APPVER := $$(shell git describe --exact-match --tag || echo dev))
	GO111MODULE=on go build -ldflags "-s -w -X main.Version=$(APPVER)"
	./upx -qq --brute tweet

clean:
	rm -f tweet

.PHONY: build
