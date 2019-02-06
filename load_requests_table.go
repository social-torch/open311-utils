/* Script to populate an AWS dynamoDB database from JSON file of Open311 Requests
If table already exists, script will add items to existing table.
If table doesn't exist, script will create table and add items

Uses credentials found in shared credentials file ~/.aws/credentials  //TODO change this to .env in makefile
and assumes these credentials have permission to create and put items in DynamoDB

Optional flags for ./load_requests_table include:
 -region string
        AWS region in which DynamoDB table should be created (default "us-east-1")
  -requestFile string
        JSON file containing list of example Open311 requests (default "./data/SchenectadyRequests.json")
  -tableName string
        Name of table in DynamoDB that will hold Requests data (default "Requests")
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

// A Request is an issue that a citizen makes agains a Service offered by a city
// see https://wiki.open311.org/GeoReport_v2/#get-service-request
type Request struct {
	ServiceRequestId  string  `json:"service_request_id"` // The unique ID of the service request created.
	Status            string  `json:"status"`             // The current status of the service request.
	StatusNotes       string  `json:"status_notes"`       // Explanation of why status was changed to current state or more details on current status than conveyed with status alone.
	ServiceName       string  `json:"service_name"`       // The human readable name of the service request type
	ServiceCode       string  `json:"service_code"`       // The unique identifier for the service request type
	Description       string  `json:"description"`        // A full description of the request or report submitted.
	AgencyResponsible string  `json:"agency_responsible"` // The agency responsible for fulfilling or otherwise addressing the service request.
	ServiceNotice     string  `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported.
	RequestedDateTime string  `json:"requested_datetime"` // The date and time when the service request was made.
	UpdatedDateTime   string  `json:"update_datetime"`    // The date and time when the service request was last modified. For requests with status=closed, this will be the date the request was closed.
	ExpectedDateTime  string  `json:"expected_datetime"`  // The date and time when the service request can be expected to be fulfilled. This may be based on a service-specific service level agreement.
	Address           string  `json:"address"`            // Human readable address or description of location.
	AddressId         string  `json:"address_id"`         // The internal address ID used by a jurisdictions master address repository or other addressing system.
	ZipCode           int32   `json:"zipcode"`            // The postal code for the location of the service request.
	Latitude          float32 `json:"lat"`                // latitude using the (WGS84) projection.
	Longitude         float32 `json:"lon"`                // longitude using the (WGS84) projection.
	MediaUrl          string  `json:"media_url"`          // A URL to media associated with the request, eg an image.
}

func main() {

	//Set up command line flags and defaults for Requests File, AWS Region, and DynamoDB table name
	requestsFilePtr := flag.String("requestFile", "./data/SchenectadyRequests.json", "JSON file containing list of example requests")
	regionPtr := flag.String("region", "us-east-1", "AWS region in which DynamoDB table should be created")
	tableNamePtr := flag.String("tableName", "Requests", "Name of table in DynamoDB that will hold Requests data")

	flag.Parse()

	// Read JSON file of Example Requests
	requests, err := readRequestsJson(*requestsFilePtr)
	if err != nil {
		fmt.Println("Error reading requests json file.")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Read " + strconv.Itoa(len(requests)) + " Requests from JSON")

	// Initialize an AWS  session in specified region that SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.  //TODO change this to use .env file
	sess, err := session.NewSession(&aws.Config{Region: aws.String(*regionPtr)})
	if err != nil {
		fmt.Println("Error creating AWS session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	svc := dynamodb.New(sess)

	// Create DynamoDB table to hold requests.  Function doesn't return until table is ready for items to be written
	_, err = createRequestsTable(svc, *tableNamePtr)
	if err != nil {
		fmt.Println("Error creating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Populate DynamoDB Requests Table with Items from JSON file
	itemsAdded, err := populateRequestsTable(svc, *tableNamePtr, requests)
	if err != nil {
		fmt.Println("Error populating '" + *tableNamePtr + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("This script added " + strconv.Itoa(itemsAdded) + " items to the '" + *tableNamePtr + "' table.")

}

// Utility function to read JSON file and unmarshal into array of Open311 Requests
func readRequestsJson(filename string) ([]Request, error) {
	raw, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var requests []Request
	err = json.Unmarshal(raw, &requests)
	if err != nil {
		fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + filename)
		fmt.Println(err.Error())
		return requests, err
	}
	return requests, err
}

// Function to create AWS DynamoDB table of given name.  Function does not return until table is ACTIVE
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
	fmt.Println("Created the table '" + tableName + "' in " + *svc.Client.Config.Region)

	return result, err
}

// Function to put items into an existing and active DynamoDB database table.  Returns the number of items added
func populateRequestsTable(svc *dynamodb.DynamoDB, tableName string, requests []Request) (int, error) {
	numItems := 0

	for _, request := range requests {
		av, err := dynamodbattribute.MarshalMap(request)

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

		// Add item in Requests table
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			return numItems, err
		}

		fmt.Println("Successfully added '" + request.ServiceRequestId + "' (" + request.ServiceName + ") to " + tableName + " table")
		numItems++
	}

	return numItems, nil
}
