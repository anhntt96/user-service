package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	tableBasics := TableBasics{TableName: "HelloCdkStack-UserTableBD4BF69E-EX0YO36MVE4Z",
		DynamoDbClient: dynamodb.NewFromConfig(cfg)}

	r := gin.Default()

	r.POST("/users", func(c *gin.Context) {
		var user User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = tableBasics.AddUser(user)
		if err != nil {
			log.Printf("Couldn't add user %v to DynamoDB. Here's why %v\n", user.email, err)
		}
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	r.GET("/users/:email", func(c *gin.Context) {
		email := c.Param("email")
		log.Print(email)
		user, err := tableBasics.GetUser(email)
		log.Printf("Fullname is %v", user.fullname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	r.Run()
}
