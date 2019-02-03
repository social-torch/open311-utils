/* Script to populate an AWS dynamoDB database from JSON file of Open311 Services
Uses .aws credentials to specify endpoint //TODO change this to .env in makefile
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io/ioutil"
	"os"
	"time"
)

// A Service is offered by a city and defines what requests a citizen can make
// Single service (type) offered via Open311
// see https://wiki.open311.org/GeoReport_v2/#get-service-list
type Service struct {
	ServiceCode string   `json:"service_code"`
	ServiceName string   `json:"service_name"`
	Description string   `json:"description"`
	Metadata    bool     `json:"metadata"`
	Type        string   `json:"type"`
	Keywords    []string `json:"keywords"` // Note: Keywords is an array
	Group       string   `json:"group"`
}

func main() {

	//Set up command line flags and defaults for Services File, AWS Region, and DynamoDB table name
	servicesFilePtr := flag.String("serviceFile", "./data/SchenectadyServices.json", "JSON file containing list of Open311 Services offered by city")
	regionPtr := flag.String("region", "us-east-1", "AWS region in which DynamoDB table should be created")
	tableNamePtr := flag.String("tableName", "Services", "Name of table in DynamoDB that will hold Services data")

	flag.Parse()

	// Read JSON file of Services Available
	services, err := readServicesJson(*servicesFilePtr)
	if err != nil {
		fmt.Println("Error reading services json file.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Initialize an AWS  session in specified region that SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{Region: aws.String(*regionPtr)})
	if err != nil {
		fmt.Println("Error creating AWS session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create DynamoDB table to hold services.  Function doesn't return until table is ready for items to be written
	svc := dynamodb.New(sess)
	_, err = createServicesTable(svc, *tableNamePtr)
	if err != nil {
		fmt.Println("Error creating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Populate DynamoDB Services Table with Items from JSON file
	itemsAdded, err := populateServicesTable(svc, *tableNamePtr, services)
	if err != nil {
		fmt.Println("Error populating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(itemsAdded)

}

func readServicesJson(filename string) ([]Service, error) {
	raw, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var services []Service
	err = json.Unmarshal(raw, &services)
	if err != nil {
		fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + filename)
		fmt.Println(err.Error())
		return nil, err
	}
	return services, nil
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
			ReadCapacityUnits:  aws.Int64(5), // TODO Determine right default provisioned capacity
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(tableName),
	}

	result, err := svc.CreateTable(input)
	if err != nil {
		return result, err
	}

	// Get current table description to see if table is ready to write items
	descInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}
	description, _ := svc.DescribeTable(descInput)

	// Block until table creation is complete
	fmt.Println("Creating table '" + tableName + "' in DynamoDB. . .")
	for *description.Table.TableStatus != "ACTIVE" {
		fmt.Println("Waiting on AWS. . . ")
		time.Sleep(5000 * time.Millisecond)
		description, _ = svc.DescribeTable(descInput)
	}
	fmt.Println("Created the table " + tableName + " in " + *svc.Client.Config.Region)

	return result, err
}

// TODO comment function and returns.
func populateServicesTable(svc *dynamodb.DynamoDB, tableName string, services []Service) (int, error) {
	numItems := 0

	for _, service := range services {
		av, err := dynamodbattribute.MarshalMap(service)

		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			return numItems, err
		}

		// Note: if item already exists, DynamoDB doesn't duplicate it

		// Prepare input for call to DynamoDB
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		// Add item in Services table
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			return numItems, err
		}

		fmt.Println("Successfully added '" + service.ServiceName + "' (" + service.Group + ") to " + tableName + " table")
		numItems++
	}

	return numItems, nil

}
