package getcollection

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	fmt.Println("Hello world!")
	collection := client.Database("myGoappDB").Collection("Posts")
	return collection

}
