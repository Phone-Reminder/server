package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AddReminder struct {
	UserID      int       `json:"userid"`
	Date        time.Time `json:"date"`
	PhoneNumber string    `json:"phonenumber"`
	Message     string    `json:"message"`
}

func main() {

	x := AddReminder{
		UserID:      4,
		Date:        time.Date(2023, 03, 9, 17, 58, 0, 0, time.UTC),
		PhoneNumber: "+447876801343",
		Message:     "Hi it worked",
	}
	fmt.Println(x)

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

	//LOGIC
	// Schedule a task to run every minute
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			currTime := time.Now().UTC()
			ctx := context.Background()
			var reminder AddReminder
			err := collection.FindOne(ctx, bson.M{"date": bson.M{"$lt": currTime}}).Decode(&reminder)

			if err != nil {
				log.Printf("Failed to find reminder: %v", err)
				continue
			}

			if reminder.Date.Before(currTime) {
				// The reminder date has passed the current time
				fmt.Println("if statement entered")
				clientSMS := twilio.NewRestClient()

				params := &api.CreateMessageParams{}
				params.SetBody(reminder.Message)
				params.SetFrom("+13203357753")
				params.SetTo(reminder.PhoneNumber)

				resp, err := clientSMS.Api.CreateMessage(params)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					if resp.Sid != nil {
						fmt.Println(*resp.Sid)
					} else {
						fmt.Println(resp.Sid)

					}

				}
				_, err = collection.DeleteOne(ctx, bson.M{"date": reminder.Date})
				if err != nil {
					log.Printf("Failed to delete reminder: %v", err)
				}
			}
		}
	}()

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

	//server
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Listening to the Localhost",
		})
	})
	r.Run("localhost:8080")
	fmt.Println("hi")
}
