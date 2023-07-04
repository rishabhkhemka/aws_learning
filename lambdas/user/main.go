package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var tableName = aws.String("users-contact-info")

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

func FetchUserViaUserID(userID string) (events.APIGatewayProxyResponse, error) {
	db := getDB()

	// create a DynamoDB get input
	input := &dynamodb.GetItemInput{
		TableName: tableName,
		Key: map[string]*dynamodb.AttributeValue{
			"userID": {
				S: aws.String(userID),
			},
		},
	}

	// perform the get operation
	result, err := db.GetItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if result.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "user not found",
		}, nil
	}

	// unmarshal the item into a User struct
	user := User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "unmarshal error from db response to struct",
		}, err
	}

	// convert the user to JSON
	responseBody, err := json.Marshal(user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "marshall error from struct to json",
		}, err
	}

	// return the response
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func CreateNewUser(body []byte) (events.APIGatewayProxyResponse, error) {
	// extract the user data from the request body
	var user User
	err := json.Unmarshal([]byte(body), &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "invalid request body",
		}, nil
	}

	// validate the required fields
	if user.UserID == "" || user.FirstName == "" || user.LastName == "" || user.Address == "" || user.MobileNumber == "" || user.EmailAddress == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "missing required fields",
		}, nil
	}

	db := getDB()

	// marshal the user object into a DynamoDB attribute map
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// create a DynamoDB put input
	input := &dynamodb.PutItemInput{
		TableName: tableName,
		Item:      item,
	}

	// perform the put operation
	_, err = db.PutItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// return success response
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "user saved successfully",
	}, nil
}

func DeleteUser(userID, firstName, lastName string) (events.APIGatewayProxyResponse, error) {
	// at least one query parameter must be provided
	if userID == "" && firstName == "" && lastName == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "provide at least one of userID, firstName, or lastName",
		}, nil
	}

	execDeleteQuery := func(input *dynamodb.DeleteItemInput) (events.APIGatewayProxyResponse, error) {
		db := getDB()

		// perform the delete operation
		_, err := db.DeleteItem(input)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		// return success response
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "user deleted successfully",
		}, nil
	}

	if userID != "" {
		input := &dynamodb.DeleteItemInput{
			TableName: tableName,
			Key: map[string]*dynamodb.AttributeValue{
				"userID": {
					S: aws.String(userID),
				},
			},
		}
		return execDeleteQuery(input)
	}

	if firstName != "" {
		users := FetchUserViaFirstName(firstName)
		if len(users) == 0 {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       "no user found for given firstName",
			}, nil
		}

		for i := range users {
			input := &dynamodb.DeleteItemInput{
				TableName: tableName,
				Key: map[string]*dynamodb.AttributeValue{
					"userID": {
						S: aws.String(users[i].UserID),
					},
				},
			}
			execDeleteQuery(input)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Body:       fmt.Sprintf("%d items deleted successfully", len(users)),
		}, nil
	}

	if lastName != "" {
		users := FetchUserViaLastName(lastName)
		if len(users) == 0 {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       "no user found for given firstName",
			}, nil
		}

		for i := range users {
			input := &dynamodb.DeleteItemInput{
				TableName: tableName,
				Key: map[string]*dynamodb.AttributeValue{
					"userID": {
						S: aws.String(users[i].UserID),
					},
				},
			}
			execDeleteQuery(input)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Body:       fmt.Sprintf("%d items deleted successfully", len(users)),
		}, nil
	}

	panic("cannot form delete query, check query params")
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

func UpdateUserByUserID(userID string, body []byte) (events.APIGatewayProxyResponse, error) {
	var user User
	err := json.Unmarshal([]byte(body), &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "invalid request body",
		}, nil
	}

	user.UserID = userID

	// validate the required fields
	if user.UserID == "" || user.FirstName == "" || user.LastName == "" || user.Address == "" || user.MobileNumber == "" || user.EmailAddress == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "missing required fields",
		}, nil
	}

	db := getDB()

	// marshal the user object into a DynamoDB attribute map
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// create a DynamoDB put input
	input := &dynamodb.PutItemInput{
		TableName: tableName,
		Item:      item,
	}

	// perform the put operation
	_, err = db.PutItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// return success response
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "user updated successfully",
	}, nil
}

func Delegator(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "GET":
		if request.PathParameters == nil {
			break
		}
		return FetchUserViaUserID(request.PathParameters["userID"])
	case "POST":
		return CreateNewUser([]byte(request.Body))
	case "DELETE":
		if request.QueryStringParameters == nil {
			break
		}
		return DeleteUser(
			request.QueryStringParameters["userID"],
			request.QueryStringParameters["firstName"],
			request.QueryStringParameters["lastName"],
		)
	case "PATCH":
		return UpdateUserByUserID(request.PathParameters["userID"], []byte(request.Body))
	}

	// unknown request
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       string("invalid path requested"),
	}, errors.New("invalid path")
}

func main() {
	lambda.Start(Delegator)
}
