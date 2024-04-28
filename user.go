package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type User struct {
	email    string `dynamodbav:"email"`
	password string `dynamodbav:"password"`
	fullname string `dynamodbav:"fullname"`
}

func (user User) GetKey() map[string]types.AttributeValue {

	email, err := attributevalue.Marshal(user.email)
	if err != nil {
		panic(err)
	}

	return map[string]types.AttributeValue{"email": email}
}

func (user User) String() string {
	return fmt.Sprintf("%v - %v ", user.email, user.fullname)
}
