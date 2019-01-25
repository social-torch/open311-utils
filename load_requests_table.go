/* Script to populate an AWS dynamoDB database from JSON file of Open311 Requests */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/social-torch/open311-util/types"
	"io/ioutil"
	"os"
)

func main() {

	//Set up command line flags and defaults for Requests File, AWS Region, and DynamoDB table name
	requestsFilePtr := flag.String("requestFile", "./SchenectadyRequests.json", "JSON file containing list of example requests")
	regionPtr := flag.String("region", "us-east-1", "AWS region in which DynamoDB table should be created")
	tableNamePtr := flag.String("tableName", "Requests", "Name of table in DynamoDB that will hold Requests data")

	flag.Parse()

	// Read JSON file of Requests
	requests := getRequestsFromFile(*requestsFilePtr)

	// Initialize an AWS  session in specified region that SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{Region: aws.String(*regionPtr)})

	if err != nil {
		fmt.Println("Error creating AWS session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create DynamoDB table to hold services
	svc := dynamodb.New(sess)
	createRequestsTable(svc, *tableNamePtr)

	// TODO block until table is finished being created.

	// Populate DynamoDB Table with Items from JSON file
	populateRequestsTable(svc, *tableNamePtr, requests)

}

func getRequestsFromFile(filename string) []open311.Request {
	raw, filErr := ioutil.ReadFile(filename)

	if filErr != nil {
		fmt.Println(filErr.Error())
		os.Exit(1)
	}

	var requests []open311.Request
	marshalErr := json.Unmarshal(raw, &requests)
	if marshalErr != nil {
		fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + filename)
		fmt.Println(marshalErr.Error())
		os.Exit(1)
	}
	return requests
}

func createRequestsTable(svc *dynamodb.DynamoDB, tableName string) (*dynamodb.CreateTableOutput, error) {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("service_request_id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("service_request_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(tableName),
	}

	result, err := svc.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceInUseException:
				fmt.Println(dynamodb.ErrCodeResourceInUseException, aerr.Error())
			case dynamodb.ErrCodeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return result, err
	}

	fmt.Println("Created the table " + tableName + " in " + *svc.Client.Config.Region)
	return result, err
}

func populateRequestsTable(svc *dynamodb.DynamoDB, tableName string, requests []open311.Request) {

	for _, request := range requests {
		av, err := dynamodbattribute.MarshalMap(request)

		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// TODO - check if item already exists?  (although, I think Dynamo might already handle this)

		// Create item in Requests table
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = svc.PutItem(input)

		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully added '" + request.ServiceRequestId + "' (" + request.ServiceName + ") to " + tableName + " table")

	}

}
