package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string `json:"-"`
}

type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *gorm.DB

func main() {
	secretName := "dev/clientID"
	region := "us-east-1"

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		panic("failed to retrieve secret from AWS Secrets Manager")
	}

	var secret Secret
	json.Unmarshal([]byte(*result.SecretString), &secret)

	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/dbname?charset=utf8&parseTime=True&loc=Local", secret.Username, secret.Password))
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.AutoMigrate(&User{})

	r := gin.Default()

	r.GET("/users", GetUsers)
	r.POST("/users", CreateUser)
	r.GET("/users/:id", GetUser)
	r.PUT("/users/:id", UpdateUser)
	r.DELETE("/users/:id", DeleteUser)

	r.Run()
}

func GetUsers(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error fetching users", "details": err.Error()})
		return
	}
	c.JSON(200, users)
}

func CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Bad request"})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	if err := db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error creating user", "details": err.Error()})
		return
	}
	c.JSON(200, user)
}

func GetUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found", "details": err.Error()})
		return
	}
	c.JSON(200, user)
}

func UpdateUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found", "details": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Bad request"})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	if err := db.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error updating user", "details": err.Error()})
		return
	}
	c.JSON(200, user)
}

func DeleteUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found", "details": err.Error()})
		return
	}
	if err := db.Delete(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error deleting user", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": "User deleted"})
}
