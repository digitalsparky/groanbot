SHELL:=/bin/bash

build: mkbindir build-tweet

mkbindir:
	(test ! -d bin && mkdir bin) || true

build-tweet:
	make -C tweet
	mv tweet/tweet bin/

deploy:
	sls deploy

clean:
	rm -rf bin
	make -C tweet clean

.PHONY: build
