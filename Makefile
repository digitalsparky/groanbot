SHELL:=/bin/bash

build: mkbindir build-tweet

mkbindir:
	(test ! -d bin && mkdir bin) || true

build-tweet:
	cd tweet && GO111MODULE=on go build -ldflags "-s -w" && ../upx -qq --brute tweet && mv tweet ../bin/

deploy:
	sls deploy

clean:
	rm -f bin/*

.PHONY: build
