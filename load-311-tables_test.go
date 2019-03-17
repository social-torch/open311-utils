package main

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// Double check table names and json fields aren't deviating from standard
const (
	testServicesTable = "Services"
	testRequestsTable = "Requests"
	testCitiesTable   = "Cities"

	testServiceCode      = "service_code"
	testServiceRequestID = "service_request_id"
	testCityName         = "city_name"
)

// Define a mock struct to be used in your unit tests
type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	// payload map[string]string // Store expected return values
	// err     error
}

// Towards checking if creating a table that already exists, start with none and keep track of added tables
var listOfTables []string

func (m *mockDynamoDBClient) CreateTable(input *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error) {

	primaryKey := *input.KeySchema[0].AttributeName
	firstAttribute := *input.AttributeDefinitions[0].AttributeName

	if primaryKey != firstAttribute {
		// TODO validate logic
		return nil, errors.New("CreateTable Testing: invalid Table Input")
	}

	tableName := *input.TableName

	output := &dynamodb.CreateTableOutput{
		TableDescription: &dynamodb.TableDescription{
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
			TableName:   aws.String(tableName),
			TableStatus: aws.String("ACTIVE"),
		},
	}

	// Check if table already exists.  If so, return ResouceInUseException
	for _, table := range listOfTables {
		if table == tableName {
			mockErr := awserr.New(dynamodb.ErrCodeResourceInUseException, "Mock ResouceInUseException Error", errors.New("Mock Error"))
			return output, mockErr
		}
	}

	// If this is a new table, keep track of it and return normally
	listOfTables = append(listOfTables, tableName)
	return output, nil
}

func (m *mockDynamoDBClient) WaitUntilTableExists(input *dynamodb.DescribeTableInput) error {
	if !testing.Short() {
		// Actual AWS takes a few seconds to finish creating the table
		time.Sleep(time.Second * 3)
	}
	return nil
}

//WaitUntilTableExists      WaitUntilTableExists(*dynamodb.DescribeTableInput) error

func TestCreateTable(t *testing.T) {

	validServicesOutput := &dynamodb.CreateTableOutput{
		TableDescription: &dynamodb.TableDescription{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String(testServiceCode),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(testServiceCode),
					KeyType:       aws.String("HASH"),
				},
			},
			TableName:   aws.String(testServicesTable),
			TableStatus: aws.String("ACTIVE"),
		},
	}
	validRequestsOutput := &dynamodb.CreateTableOutput{
		TableDescription: &dynamodb.TableDescription{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String(testServiceRequestID),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(testServiceRequestID),
					KeyType:       aws.String("HASH"),
				},
			},
			TableName:   aws.String(testRequestsTable),
			TableStatus: aws.String("ACTIVE"),
		},
	}
	validCitiesOutput := &dynamodb.CreateTableOutput{
		TableDescription: &dynamodb.TableDescription{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String(testCityName),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(testCityName),
					KeyType:       aws.String("HASH"),
				},
			},
			TableName:   aws.String(testCitiesTable),
			TableStatus: aws.String("ACTIVE"),
		},
	}

	tests := []struct {
		name       string
		tableName  string
		primaryKey string
		wantResult *dynamodb.CreateTableOutput
		wantErr    error
	}{
		{"Valid Services table", ServicesTable, "service_code", validServicesOutput, nil},
		{"Valid Requests table", RequestsTable, "service_request_id", validRequestsOutput, nil},
		{"Valid Cities table", CitiesTable, "city_name", validCitiesOutput, nil},
		{"Table already exists", ServicesTable, "service_code", validServicesOutput, nil},
	}

	// Create mock dynamo service client
	mockSvc := &mockDynamoDBClient{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createTable(mockSvc, tt.tableName, tt.primaryKey)

			if result == nil {
				t.Fatalf("createTable should not return nil")
			}
			if err != tt.wantErr {
				t.Errorf("want error %s; got %s", tt.wantErr, err)
			}
			if *result.TableDescription.AttributeDefinitions[0].AttributeName != *tt.wantResult.TableDescription.AttributeDefinitions[0].AttributeName {
				t.Errorf("want index %s; got %s", *tt.wantResult.TableDescription.AttributeDefinitions[0].AttributeName, *result.TableDescription.AttributeDefinitions[0].AttributeName)
			}
			if *result.TableDescription.KeySchema[0].AttributeName != *tt.wantResult.TableDescription.KeySchema[0].AttributeName {
				t.Errorf("want key %s; got %s", *tt.wantResult.TableDescription.KeySchema[0].AttributeName, *result.TableDescription.KeySchema[0].AttributeName)
			}
			if *result.TableDescription.TableName != *tt.wantResult.TableDescription.TableName {
				t.Errorf("want table name %s; got %s", *tt.wantResult.TableDescription.TableName, *result.TableDescription.TableName)
			}
			if *result.TableDescription.TableStatus != *tt.wantResult.TableDescription.TableStatus {
				t.Errorf("want table status %s; got %s", *tt.wantResult.TableDescription.TableStatus, *result.TableDescription.TableStatus)
			}

		})
	}

}
