#!/bin/bash
cd tweet
go build -ldflags "-s -w"
mv tweet ../bin/
