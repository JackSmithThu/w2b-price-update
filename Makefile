SHELL := /bin/bash

all: clean format output build 

format:
	find . -name '*.go' | grep -Ev 'vendor|thrift_gen|proto_gen' | xargs gofmt -w
output:
	mkdir -p output/conf
	cp -rfv conf/* output/conf
	cp -rfv script/* output/
	chmod 755 output/bootstrap.sh
build:
	mkdir -p output/bin 
	go build -i -o output/bin/ares.script.gt_price_update
	chmod 755 output/bin/ares.script.gt_price_update

clean:
	rm -rf output

run:
	./output/bootstrap.sh

server: all run
