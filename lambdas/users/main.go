package main

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var tableName = aws.String(os.Getenv("TABLE_NAME"))

func getDB() *dynamodb.DynamoDB {
	sess, err := session.NewSession()
	if err != nil {
		panic("could not create aws session")
	}

	dynamoDBClient := dynamodb.New(sess)
	return dynamoDBClient
}

type User struct {
	UserID       string `json:"userID"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Address      string `json:"address"`
	MobileNumber string `json:"mobileNumber"`
	EmailAddress string `json:"emailAddress"`
}

func FetchAllUsers() []User {
	db := getDB()

	// Create a DynamoDB scan input
	input := &dynamodb.ScanInput{
		TableName: tableName,
	}

	// Perform the scan operation
	result, err := db.Scan(input)
	if err != nil {
		return nil
	}

	// Convert the DynamoDB items to a list of users
	users := make([]User, 0)
	for _, item := range result.Items {
		user := User{}
		err = dynamodbattribute.UnmarshalMap(item, &user)
		if err != nil {
			return nil
		}
		users = append(users, user)
	}
	return users
}

func FetchUserViaLastName(lastName string) []User {
	db := getDB()

	queryInput := &dynamodb.QueryInput{
		TableName:              tableName,
		IndexName:              aws.String("lastNameIndex"),
		KeyConditionExpression: aws.String("lastName = :ln"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":ln": {S: aws.String(lastName)},
		},
	}

	queryResult, err := db.Query(queryInput)
	if err != nil {
		return nil
	}

	users := []User{}
	for i := range queryResult.Items {
		item := queryResult.Items[i]
		user := User{}
		err = dynamodbattribute.UnmarshalMap(item, &user)
		if err != nil {
			return nil
		}
		users = append(users, user)
	}

	return users
}

func FetchUserViaFirstName(firstName string) []User {
	db := getDB()

	queryInput := &dynamodb.QueryInput{
		TableName:              tableName,
		IndexName:              aws.String("firstNameIndex"),
		KeyConditionExpression: aws.String("firstName = :fn"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":fn": {S: aws.String(firstName)},
		},
	}

	queryResult, err := db.Query(queryInput)
	if err != nil {
		return nil
	}

	users := []User{}
	for i := range queryResult.Items {
		item := queryResult.Items[i]
		user := User{}
		err = dynamodbattribute.UnmarshalMap(item, &user)
		if err != nil {
			return nil
		}
		users = append(users, user)
	}

	return users
}

func Delegator(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.QueryStringParameters == nil {
		users := FetchAllUsers()
		return convertToResponse(users)
	}

	if request.QueryStringParameters["sorted"] == "true" {
		users := FetchAllUsers()
		sort.Slice(users, func(i, j int) bool {
			if users[i].FirstName < users[j].FirstName {
				return true
			}
			if users[i].FirstName == users[j].FirstName {
				return users[i].LastName < users[j].LastName
			}
			return false
		})
		return convertToResponse(users)
	}

	if name := request.QueryStringParameters["name"]; name != "" {
		byFirstName := FetchUserViaFirstName(name)
		byLastName := FetchUserViaLastName(name)
		users := []User{}
		users = append(users, byFirstName...)
		users = append(users, byLastName...)
		return convertToResponse(users)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       "cannot resolve query",
	}, nil
}

func convertToResponse(users []User) (events.APIGatewayProxyResponse, error) {
	if len(users) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "no user found with given queries",
		}, nil
	}
	bytes, err := json.Marshal(users)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "could not marshal list of users",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}, nil
}

func main() {
	lambda.Start(Delegator)
}
