.PHONY: build install

build:
	go get -u github.com/elanq/wakawaka
	go build

install:
	go install
