#include .env

build: load-311-tables

clean:
	@echo "  >  Cleaning up..."
	go clean

dep:
	@echo "  >  Getting dependencies..."
	go get github.com/aws/aws-sdk-go
	go get flag

load-311-tables:
	@echo "  >  Building binary..."
	go build load-311-tables.go

deploy: build
	@echo "  >  Creating and populating DynamoDB tables for Services, Requests, and Cities"
	./load-311-tables \
        --serviceFile ./data/SchenectadyServices.json \
        --requestFile ./data/SchenectadyRequests.json \
        --cityFile ./data/Cities.json \
        --region "us-east-1"


.PHONY: clean dep build
