package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AddReminder struct {
	UserID      int    `json:"userid"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
	Message     string `json:"message"`
}

func main() {

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connected to MongoDB!")

	// Get a handle for your collection
	collection := client.Database("reminder").Collection("reminders")

	r := gin.Default()
	r.POST("/addReminder", func(c *gin.Context) {
		var addRemData AddReminder
		if err := c.ShouldBindJSON(&addRemData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Insert the reminder into the collection
		_, err = collection.InsertOne(context.TODO(), addRemData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Reminder added Successfully."})
	})
	r.GET("/getReminder", func(c *gin.Context) {
		// Find all reminders in the collection
		// Find all the reminders
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Define a slice to hold the reminders
	var reminders []AddReminder

	// Iterate over the cursor and decode the reminders into the slice
	for cur.Next(context.TODO()) {
		var reminder AddReminder
		err := cur.Decode(&reminder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reminders = append(reminders, reminder)
	}

	// Return the slice of reminders
	c.JSON(http.StatusOK, gin.H{"reminders": reminders})
})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Listening to the Localhost",
		})
	})
	r.Run("localhost:8080")
}
