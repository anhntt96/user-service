package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-west-2"))

func main() {
	r := gin.Default()

	r.POST("/users", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		createUser(user)
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := getUser(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	r.PUT("/users/:id", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := updateUser(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	r.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := deleteUser(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "User deleted!"})
	})

	r.Run()
}

// Create a user
func createUser(user User) {
	av, _ := dynamodbattribute.MarshalMap(user)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Users"),
	}
	_, _ = db.PutItem(input)
}

// Get a user
func getUser(id string) (*User, error) {
	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	user := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, user)
	return user, err
}

// Update a user
func updateUser(user User) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":n": {
				S: aws.String(user.Name),
			},
			":e": {
				S: aws.String(user.Email),
			},
		},
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(user.ID),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set Name = :n, Email = :e"),
	}

	_, err := db.UpdateItem(input)
	return err
}

// Delete a user
func deleteUser(id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	}

	_, err := db.DeleteItem(input)
	return err
}
