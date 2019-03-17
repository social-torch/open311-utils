/* Script to populate an AWS dynamoDB database from JSON files of Open311 Services, Requests, and Cities
If tables already exist, script will add items to existing table.
If tables don't exist, script will create tables and add items

Uses credentials found in shared credentials file ~/.aws/credentials
and assumes these credentials have permission to create and put items in DynamoDB

Usage of ./load-311-tables:
  -serviceFile string
    	JSON file containing list of Open311 Services offered by city (default "")
  -requestFile string
    	JSON file containing list of example Open311 requests (default "")
  -cityFile string
    	JSON file containing list of cities and corresponding endpoints (default "")
  -region string
    	AWS region in which DynamoDB table should be created (default "us-east-1")

if one of serviceFile, requestFile, or cityFile is set to "", that table will not be created nor populated

note: DynamoDB tables are created with a provisioned throughput set low to qualify for AWS free tier.
       Highly utilized production instances will need to increase this

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	ServicesTable = "Services"
	RequestsTable = "Requests"
	CitiesTable   = "Cities"
)

// A Service is offered by a city and defines what requests a citizen can make
// Single service (type) offered via Open311
// see https://wiki.open311.org/GeoReport_v2/
type Service struct {
	ServiceCode string   `json:"service_code"`
	ServiceName string   `json:"service_name"`
	Description string   `json:"description"`
	Metadata    bool     `json:"metadata"`
	Type        string   `json:"type"`
	Keywords    []string `json:"keywords"`
	Group       string   `json:"group"`
}

// Possible value for ServiceAttribute that defines lists
type AttributeValue struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Issues that have been reported as service requests.  Location is submitted via lat/long or address
type Request struct {
	ServiceRequestId  string           `json:"service_request_id"` // The unique ID of the service request created.
	Status            string           `json:"status"`             // The current status of the service request.
	StatusNotes       string           `json:"status_notes"`       // Explanation of why status was changed to current state or more details on current status than conveyed with status alone.
	ServiceName       string           `json:"service_name"`       // The human readable name of the service request type
	ServiceCode       string           `json:"service_code"`       // The unique identifier for the service request type
	Description       string           `json:"description"`        // A full description of the request or report submitted.
	AgencyResponsible string           `json:"agency_responsible"` // The agency responsible for fulfilling or otherwise addressing the service request.
	ServiceNotice     string           `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported.
	RequestedDateTime string           `json:"requested_datetime"` // The date and time when the service request was made.
	UpdatedDateTime   string           `json:"update_datetime"`    // The date and time when the service request was last modified. For requests with status=closed, this will be the date the request was closed.
	ExpectedDateTime  string           `json:"expected_datetime"`  // The date and time when the service request can be expected to be fulfilled. This may be based on a service-specific service level agreement.
	Address           string           `json:"address"`            // Human readable address or description of location.
	AddressId         string           `json:"address_id"`         // The internal address ID used by a jurisdictions master address repository or other addressing system.
	ZipCode           int32            `json:"zipcode"`            // The postal code for the location of the service request.
	Latitude          float32          `json:"lat"`                // latitude using the (WGS84) projection.
	Longitude         float32          `json:"lon"`                // longitude using the (WGS84) projection.
	MediaUrl          string           `json:"media_url"`          // A URL to media associated with the request, eg an image.
	Values            []AttributeValue `json:"values"`             // Enables future expansion
}

// Assumes each jurisdiction has its own AWS endpoint
type City struct {
	CityName string `json:"city_name"`
	Endpoint string `json:"endpoint"`
}

func main() {

	//Set up command line flags and defaults for json files and AWS Region
	servicesFilePtr := flag.String("serviceFile", "", "JSON file containing list of Open311 Services offered by city")
	requestsFilePtr := flag.String("requestFile", "", "JSON file containing list of example Open311 requests")
	citiesFilePtr := flag.String("cityFile", "", "JSON file containing list of cities and corresponding endpoints")
	regionPtr := flag.String("region", endpoints.UsEast1RegionID, "AWS region in which DynamoDB table should be created")

	flag.Parse()

	if *servicesFilePtr == "" && *requestsFilePtr == "" && *citiesFilePtr == "" {
		fmt.Println("Please specify at least one JSON file to load.  For usage, type  ./load-311-tables --help")
		os.Exit(1)
	}

	// Initialize an AWS  session in specified region that SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.  //TODO change this to use .env file
	svc, err := createDynamoClient(*regionPtr)
	if err != nil {
		fmt.Println("Error creating AWS session:")
		os.Exit(1)
	}
	fmt.Println("Established AWS session in " + *svc.Client.Config.Region + "\n")

	// ///////////// Services Table  /////////////////////////////
	// Read JSON file of Services Available
	if *servicesFilePtr != "" {
		raw, err := ioutil.ReadFile(*servicesFilePtr)
		if err != nil {
			fmt.Println("Error opening services json file: " + *servicesFilePtr)
			os.Exit(1)
		}

		var services []Service
		err = json.Unmarshal(raw, &services)
		if err != nil {
			fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + *servicesFilePtr)
			os.Exit(1)
		}
		fmt.Println("Read " + strconv.Itoa(len(services)) + " Services from JSON")

		// Create DynamoDB table to hold services.  Function doesn't return until table is ready for items to be written
		result, err := createTable(svc, ServicesTable, "service_code")
		if err != nil {
			fmt.Println("Error creating '" + ServicesTable + "' table.")
			fmt.Printf("With AWS response: %+v", result)
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Populate DynamoDB Services Table with Items from JSON file
		itemsAdded, err := populateServicesTable(svc, ServicesTable, services)
		if err != nil {
			fmt.Println("Error populating '" + ServicesTable + "' table.")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Added " + strconv.Itoa(itemsAdded) + " items to the '" + ServicesTable + "' table.\n")

	}
	// ///////////// Requests Table  /////////////////////////////
	// Read JSON file of Example Requests
	if *requestsFilePtr != "" {
		raw, err := ioutil.ReadFile(*requestsFilePtr)
		if err != nil {
			fmt.Println("Error opening requests json file: " + *requestsFilePtr)
			os.Exit(1)
		}

		var requests []Request
		err = json.Unmarshal(raw, &requests)
		if err != nil {
			fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + *requestsFilePtr)
			os.Exit(1)
		}
		fmt.Println("Read " + strconv.Itoa(len(requests)) + " Requests from JSON")

		// Create DynamoDB table to hold requests.  Function doesn't return until table is ready for items to be written
		result, err := createTable(svc, RequestsTable, "service_request_id")
		if err != nil {
			fmt.Println("Error creating '" + RequestsTable + "' table.")
			fmt.Printf("With AWS response: %+v", result)
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Populate DynamoDB Requests Table with Items from JSON file
		itemsAdded, err := populateRequestsTable(svc, RequestsTable, requests)
		if err != nil {
			fmt.Println("Error populating '" + RequestsTable + "' table.")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Added " + strconv.Itoa(itemsAdded) + " items to the '" + RequestsTable + "' table.\n")

	}

	// ///////////// Cities Table  /////////////////////////////
	// Read JSON file of Example Cities and endpoints
	if *citiesFilePtr != "" {
		raw, err := ioutil.ReadFile(*citiesFilePtr)
		if err != nil {
			fmt.Println("Error opening cities json file: " + *citiesFilePtr)
			os.Exit(1)
		}

		var cities []City
		err = json.Unmarshal(raw, &cities)
		if err != nil {
			fmt.Println("Error Unmarshaling JSON.  Check Syntax in " + *citiesFilePtr)
			os.Exit(1)
		}
		fmt.Println("Read " + strconv.Itoa(len(cities)) + " Cities from JSON")

		// Create DynamoDB table to hold cities.  Function doesn't return until table is ready for items to be written
		result, err := createTable(svc, CitiesTable, "city_name")
		if err != nil {
			fmt.Println("Error creating '" + CitiesTable + "' table.")
			fmt.Printf("With AWS response: %+v", result)
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Populate DynamoDB Services Table with Items from JSON file
		itemsAdded, err := populateCitiesTable(svc, CitiesTable, cities)
		if err != nil {
			fmt.Println("Error populating '" + CitiesTable + "' table.")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Added " + strconv.Itoa(itemsAdded) + " items to the '" + CitiesTable + "' table.\n")

	}
}

// createDynamoClient is a convenience function to establish a session with AWS and
// returns a new instance of the DynamoDB client
func createDynamoClient(region string) (*dynamodb.DynamoDB, error) {

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role.

	// Create the session that the DynamoDB service will use.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, fmt.Errorf("\n repository: unable to establish session with AWS \n  %s", err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	return svc, nil
}

// Function to create AWS DynamoDB table of given name.  Function does not return until table is ACTIVE
// ensure that primaryKey passed in matches the intended JSON field of the struct
// Table is created with a provisioned throughput set low to qualify for AWS free tier.
//   Highly utilized production instances will need to increase this
//func createTable(svc *dynamodb.DynamoDB, tableName string, primaryKey string) (*dynamodb.CreateTableOutput, error) {
func createTable(svc dynamodbiface.DynamoDBAPI, tableName string, primaryKey string) (*dynamodb.CreateTableOutput, error) {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(primaryKey),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(primaryKey),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5), //Intentionally set low to qualify for AWS free tier
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(tableName),
	}

	result, err := svc.CreateTable(input)

	// If table already exists, continue gracefully
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
		} else {
			// Catch non-awsErrs and process ... highly unlikly code will execute this path
			return result, err
		}
	}

	fmt.Println("Creating table '" + tableName + "' in DynamoDB.  Waiting on AWS. . .")

	// Block until table creation is complete
	err = svc.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		fmt.Println("Error waiting on table creation for '" + tableName + "' table.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Created table: '" + tableName + "'")

	return result, err
}

// Function to put Service items into an existing and active DynamoDB database table.  Returns the number of items added
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

// Function to put Request items into an existing and active DynamoDB database table.  Returns the number of items added
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

// Function to put City items into an existing and active DynamoDB database table.  Returns the number of items added
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

		// Add item in City table
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			return numItems, err
		}

		fmt.Println("Successfully added '" + city.CityName + "' to " + tableName + " table")
		numItems++
	}

	return numItems, nil
}
