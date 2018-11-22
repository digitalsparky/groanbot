SHELL:=/bin/bash

build: mkbindir build-tweet

mkbindir:
	(test ! -d bin && mkdir bin) || true

build-tweet:
	cd tweet && go build -ldflags "-s -w" && mv tweet ../bin/

deploy:
	sls deploy

clean:
	rm -f bin/*

.PHONY: build