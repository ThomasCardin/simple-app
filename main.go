package main

import (
	"context"
	"fmt"
	"net/http"
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
	FirstName string `bson:"FirstName,omitempty"`
	LastName  string `bson:"LastName,omitempty"`
	Email     string `bson:"Email,omitempty"`
}

func main() {
	mongoUser := os.Getenv(MONGO_USERNAME)
	mongoPassword := os.Getenv(MONGO_PASSWORD)
	mongoHost := os.Getenv(MONGO_HOST)
	mongoDbUri := fmt.Sprintf("mongodb://%s:%s@%s:27017", mongoUser, mongoPassword, mongoHost)
	//mongoDbUri := "mongodb://root:root@localhost:27017"

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
		firstName := c.PostForm("FirstName")
		lastName := c.PostForm("LastName")
		email := c.PostForm("Email")

		user := User{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
		}

		_, err := collection.InsertOne(ctx, user)
		if err != nil {
			fmt.Printf("error creating the user: %s", err.Error())
		}

		c.Redirect(http.StatusMovedPermanently, "/")
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
		User{FirstName: "Cat's Cradle", LastName: "Kurt Vonnegut Jr.", Email: "Testing@allo.com"},
		User{FirstName: "In Memory of Memory", LastName: "Maria Stepanova", Email: "Testing@allo.com"},
		User{FirstName: "Pride and Prejudice", LastName: "Jane Austen", Email: "Testing@allo.com"},
	}

	_, err = usersCollection.InsertMany(ctx, users)
	if err != nil {
		fmt.Printf("error creating default users %s", err.Error())
	}

	return usersCollection
}
