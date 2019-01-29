/* Script to populate an AWS dynamoDB database from JSON file of Open311 Services */

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
	"time"
)

func main() {

	//Set up command line flags and defaults for Services File, AWS Region, and DynamoDB table name
	servicesFilePtr := flag.String("serviceFile", "./data/SchenectadyServices.json", "JSON file containing list of Open311 Services offered by city")
	regionPtr := flag.String("region", "us-east-1", "AWS region in which DynamoDB table should be created")
	tableNamePtr := flag.String("tableName", "Services", "Name of table in DynamoDB that will hold Services data")

	flag.Parse()

	// Read JSON file of Services Available
	services := getServicesFromFile(*servicesFilePtr)

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
	createServicesTable(svc, *tableNamePtr)

	// Populate DynamoDB Services Table with Items from JSON file
	populateServicesTable(svc, *tableNamePtr, services)

}

func getServicesFromFile(filename string) []open311.Service {
	raw, filErr := ioutil.ReadFile(filename)

	if filErr != nil {
		fmt.Println(filErr.Error())
		os.Exit(1)
	}

	var services []open311.Service
	marshalErr := json.Unmarshal(raw, &services)
	if marshalErr != nil {
		fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + filename)
		fmt.Println(marshalErr.Error())
		os.Exit(1)
	}
	return services
}

func createServicesTable(svc *dynamodb.DynamoDB, tableName string) (*dynamodb.CreateTableOutput, error) {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("service_code"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("service_code"),
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

	// Block until table creation is complete
	descInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}
	description, _ := svc.DescribeTable(descInput)
	for *description.Table.TableStatus != "ACTIVE" {
		fmt.Println("Table Creation Pending. Waiting on AWS. . . ")
		time.Sleep(5000 * time.Millisecond)
		description, _ = svc.DescribeTable(descInput)

	}
	fmt.Println("Created the table " + tableName + " in " + *svc.Client.Config.Region)
	return result, err
}

func populateServicesTable(svc *dynamodb.DynamoDB, tableName string, services []open311.Service) {

	for _, service := range services {
		av, err := dynamodbattribute.MarshalMap(service)

		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// TODO - check if item already exists?  (although, I think Dynamo might already handle this)

		// Create item in table Services
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

		fmt.Println("Successfully added '" + service.ServiceName + "' (" + service.Group + ") to " + tableName + " table")

	}

}
