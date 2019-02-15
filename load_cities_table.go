/* Script to populate an AWS dynamoDB database from JSON file of Cities and corresponding Open311 Endpoints
If table already exists, script will add items to existing table.
If table doesn't exist, script will create table and add items

Uses credentials found in shared credentials file ~/.aws/credentials  //TODO change this to .env in makefile
and assumes these credentials have permission to create and put items in DynamoDB

Optional flags for ./load_services_table include:
 -region string
        AWS region in which DynamoDB table should be created (default "us-east-1")
  -cityFile string
        JSON file containing list of Open311 Cities and Endpoints (default "./data/Cities.json")
  -tableName string
        Name of table in DynamoDB that will hold Services data (default "Cities")
*/

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
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// Assumes each jurisdiction has its own AWS endpoint
// CityNames need to be unique.  use "Cityname, StateCode"
type City struct {
	CityName string `json:"city_name"`
	Endpoint string `json:"endpoint"`
}

func main() {

	//Set up command line flags and defaults for Services File, AWS Region, and DynamoDB table name
	servicesFilePtr := flag.String("cityFile", "./data/Cities.json", "JSON file containing list of cities and corresponding endpoints")
	regionPtr := flag.String("region", "us-east-1", "AWS region in which DynamoDB table should be created")
	tableNamePtr := flag.String("tableName", "Cities", "Name of table in DynamoDB that will hold City data")

	flag.Parse()

	// Read JSON file of Services Available
	services, err := readServicesJson(*servicesFilePtr)
	if err != nil {
		fmt.Println("Error reading services json file.")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Read " + strconv.Itoa(len(services)) + " Services from JSON")

	// Initialize an AWS  session in specified region that SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.  //TODO change this to use .env file
	sess, err := session.NewSession(&aws.Config{Region: aws.String(*regionPtr)})
	if err != nil {
		fmt.Println("Error creating AWS session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	svc := dynamodb.New(sess)

	// Create DynamoDB table to hold services.  Function doesn't return until table is ready for items to be written
	_, err = createCitiesTable(svc, *tableNamePtr)
	if err != nil {
		fmt.Println("Error creating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Populate DynamoDB Services Table with Items from JSON file
	itemsAdded, err := populateCitiesTable(svc, *tableNamePtr, services)
	if err != nil {
		fmt.Println("Error populating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("This script added " + strconv.Itoa(itemsAdded) + " items to the '" + *tableNamePtr + "' table.")

}

// Utility function to read JSON file and unmarshal into array of Services
func readServicesJson(filename string) ([]City, error) {
	raw, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var cities []City
	err = json.Unmarshal(raw, &cities)
	if err != nil {
		fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + filename)
		fmt.Println(err.Error())
		return cities, err
	}
	return cities, err
}

// Function to create AWS DynamoDB table of given name.  Function does not return until table is ACTIVE
func createCitiesTable(svc *dynamodb.DynamoDB, tableName string) (*dynamodb.CreateTableOutput, error) {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("city_name"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("city_name"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5), //TODO Determine right default provisioned capacity
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(tableName),
	}

	result, err := svc.CreateTable(input)

	// If table already exists, return gracefully
	// see: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/handling-errors.html
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case dynamodb.ErrCodeResourceInUseException: // AWS returns ResouceInUseException to a createTable call if table already exists
				fmt.Println("Warning: '" + tableName + "' table already exists.  Continuing to add items to existing table.")
				return result, nil // If table already exists, return without error to continue
			default: // Process error generically
				return result, err
			}
		}
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

// Function to put items into an existing and active DynamoDB database table.  Returns the number of items added
func populateCitiesTable(svc *dynamodb.DynamoDB, tableName string, cities []City) (int, error) {
	numItems := 0

	for _, city := range cities {
		av, err := dynamodbattribute.MarshalMap(city)

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

		fmt.Println("Successfully added '" + city.CityName + " to " + tableName + " table")
		numItems++
	}

	return numItems, nil
}
