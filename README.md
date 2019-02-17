# Open311 Utilities
This golang utility will create and load AWS DynamoDB tables with Open311 Services, Requests, and City endpoints
for pre-populating services in a new Open311 deployment or for testing with a set of requests.



## Dependencies
```bash
$ > go get github.com/aws/aws-sdk-go
$ > go get flag
```
Also depends on AWS credentials set up as noted here:
     https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html


## Checkout
```bash
$ > mkdir -p $GOPATH/src/github.com/social-torch
$ > cd  $GOPATH/src/github.com/social-torch
$ > git clone git@github.com:social-torch/open311-utils
```
## Build
```bash
$ > go build load-311-tables.go 
```

## Usage
```bash
$ > ./load-311-tables --serviceFile ./data/SchenectadyServices.json --requestFile ./data/SchenectadyRequests.json --cityFile ./data/Cities.json --region "us-east-1"
```

##### Command line flags:
```  
  --serviceFile string
    	JSON file containing list of Open311 Services offered by city (default "./data/SchenectadyServices.json")
  --requestFile string
    	JSON file containing list of example Open311 requests (default "")
  --cityFile string
    	JSON file containing list of cities and corresponding endpoints (default "./data/Cities.json")
  --region string
    	AWS region in which DynamoDB table should be created (default "us-east-1")
```
If any of serviceFile, requestFile, or cityFile is not specified, that table will not be created nor populated.
For example, to create and load only a Services table in us-west-2 (and not create/load Requests and Cities tables):
```
$ > ./load_311_tables --serviceFile ./data/ChicagoServices.json --region "us-west-2"
```

## Makefile (optional)
A makefile really isn't necessary here (see the [go command Motivation](https://golang.org/doc/articles/go_command.html)).
But for you makefile fans, the following commands will accomplish everything listed aboveL
```bash
# Get Dependencies
$ > make dep

# Build binary
$ > make build

# Create and populate Services, Request, and Cities example tables in AWS
$ > make deploy
```


