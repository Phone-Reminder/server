package main

import (
	"context"
	"fmt"
	"log"

	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AddReminder struct {
	Date        time.Time `json:"date"`
	PhoneNumber string    `json:"phonenumber"`
	Message     string    `json:"message"`
}

func main() {

	gin.SetMode(gin.ReleaseMode)
	// Load environment variables from .env file

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	// Set client options
	dbConnection := os.Getenv("DBURI")
	fromPhoneNo := os.Getenv("FROMPHONENO")
	localHost := os.Getenv("LOCALHOST")
	twilioSID := os.Getenv("TWILIO_SID")
	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	clientOptions := options.Client().ApplyURI(dbConnection)

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
	r.Use(cors())

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
			// clientSMS := twilio.NewRestClientWithParams(twilio.ClientParams{
			// 		Username: twilioSID,
			// 		Password: twilioAuthToken,
			// 	})

			if reminder.Date.Before(currTime) {
				// The reminder date has passed the current time
				// clientSMS := twilio.NewRestClient()
				clientSMS := twilio.NewRestClientWithParams(twilio.ClientParams{
					Username: twilioSID,
					Password: twilioAuthToken,
				})

				params := &api.CreateMessageParams{}
				params.SetBody(reminder.Message)
				params.SetFrom(fromPhoneNo)
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
	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Listening to the Localhost",
	// 	})
	// })

	r.GET("/", func(c *gin.Context) {
		url := "http://127.0.0.1:3000"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header = c.Request.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})
	r.Run(localHost)

}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://reminders-bt-ss.netlify.app/")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
