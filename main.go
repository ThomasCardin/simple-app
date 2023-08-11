package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	MONGO_USERNAME = "MONGO_USERNAME"
	MONGO_PASSWORD = "MONGO_PASSWORD"
	MONGO_HOST     = "MONGO_HOST"

	SERVER_PORT = "SERVER_PORT"
)

type User struct {
	FistName string `bson:"FistName,omitempty"`
	LastName string `bson:"LastName,omitempty"`
	Email    string `bson:"Email,omitempty"`
}

func main() {
	mongoUser := os.Getenv(MONGO_USERNAME)
	mongoPassword := os.Getenv(MONGO_PASSWORD)
	mongoHost := os.Getenv(MONGO_HOST)
	mongoDbUri := fmt.Sprintf("mongodb://%s:%s@%s:27017", mongoUser, mongoPassword, mongoHost)

	ctx := context.Background()
	collection := SetUpDatabase(ctx, mongoDbUri)

	r := gin.Default()
	r.LoadHTMLGlob("index.html")

	r.GET("/", func(c *gin.Context) {
		cursor, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			fmt.Printf("error getting the collection: %s", err.Error())
			defer cursor.Close(ctx)
		}

		var users []User
		for cursor.Next(ctx) {
			var result User
			err := cursor.Decode(&result)
			if err != nil {
				fmt.Printf("error decoding documents: %s", err.Error())
			}
			users = append(users, result)
		}

		c.HTML(200, "index.html", gin.H{
			"users": users,
		})
	})

	r.POST("/add", func(c *gin.Context) {
		firstName := c.PostForm("FistName")
		lastName := c.PostForm("LastName")
		email := c.PostForm("Email")

		user := User{
			FistName: firstName,
			LastName: lastName,
			Email:    email,
		}

		_, err := collection.InsertOne(ctx, user)
		if err != nil {
			fmt.Printf("error creating the user: %s", err.Error())
		}

		c.String(200, fmt.Sprintf("User added : %+v", user))
	})

	r.Run()
}

func SetUpDatabase(ctx context.Context, mongoDbUri string) *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoDbUri))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Printf("Error pinging the database")
	}

	usersCollection := client.Database("statefull-go-app").Collection("users")
	users := []interface{}{
		User{FistName: "Cat's Cradle", LastName: "Kurt Vonnegut Jr.", Email: "Testing@allo.com"},
		User{FistName: "In Memory of Memory", LastName: "Maria Stepanova", Email: "Testing@allo.com"},
		User{FistName: "Pride and Prejudice", LastName: "Jane Austen", Email: "Testing@allo.com"},
	}

	_, err = usersCollection.InsertMany(ctx, users)
	if err != nil {
		fmt.Printf("error creating default users %s", err.Error())
	}

	return usersCollection
}
