# Open311 Utilities

```bash
$ > mkdir -p $GOPATH/src/github.com/social-torch
$ > cd  $GOPATH/src/github.com/social-torch
$ > git clone git@github.com:social-torch/open311-utils
$ > cd open311-utils

```

## Dependencies
```bash
$ > go get flag
$ > go get github.com/aws/aws-sdk-go
```

## Build
```bash
$ > go build load_services_table.go 
$ > go build load_requests_table.go 
```

## Usage

```bash
$ > ./load_services_table
$ > ./load_requests_table
```

Usage of ./load_services_table:
|Command line flag | Description | Default|
|-------------------|-------------|--------|
|-region|  AWS region in which DynamoDB table should be created| "us-east-1" |
|-serviceFile |JSON file containing list of Open311 Services offered by city | ""./data/SchenectadyServices.json" |
|-tableName | Name of table in DynamoDB that will hold Services data | "Services"|

Usage of ./load_requests_table:
|Command line flag | Description | Default|
|-------------------|-------------|--------|
|-region|  AWS region in which DynamoDB table should be created| "us-east-1" |
|-requestFile | JSON file containing list of example requests | "./data/SchenectadyRequests.json" |
|-tableName | Name of table in DynamoDB that will hold Requests data | "Requests" |



### Specify command line options

```bash
$ > ./load_services_table --region "us-east-1" --serviceFile "./data/ChicagoServices.json" --tableName "ChiServices"
$ > ./load_requests_table --region "us-east-1" --requestFile "./data/ChicagoRequests.json" --tableName "ChiRequests"
```