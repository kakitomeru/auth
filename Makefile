SHELL := /bin/bash

start:
	(set -a; source .env; set +a; go run cmd/main.go)

update-pkg:
	go get -u github.com/kakitomeru/shared@main
	go mod tidy