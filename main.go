package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type User struct {
	FistName string `bson:"FistName,omitempty"`
	LastName string `bson:"LastName,omitempty"`
	Email    string `bson:"Email,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	mongoDbUri := os.Getenv("MONGODB_URI")

	ctx := context.Background()
	collection := SetUpDatabase(ctx, mongoDbUri)

	r := gin.Default()
	r.LoadHTMLGlob("*.html") // Assurez-vous que votre fichier HTML est dans le même répertoire

	r.GET("/", func(c *gin.Context) {
		cursor, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Fatalf("error getting the collection: %s", err.Error())
			defer cursor.Close(ctx)
		}

		var users []User
		for cursor.Next(ctx) {
			var result User
			err := cursor.Decode(&result)
			if err != nil {
				log.Fatalf("error decoding documents: %s", err.Error())
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
			log.Fatalf("error creating the user: %s", err.Error())
		}

		c.String(200, fmt.Sprintf("User added : %+v", user))
	})

	r.Run(fmt.Sprintf(":%s", port))
}

func SetUpDatabase(ctx context.Context, mongoDbUri string) *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoDbUri))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("Error pinging the database")
	}

	usersCollection := client.Database("statefull-go-app").Collection("users")
	users := []interface{}{
		User{FistName: "Cat's Cradle", LastName: "Kurt Vonnegut Jr.", Email: "Testing@allo.com"},
		User{FistName: "In Memory of Memory", LastName: "Maria Stepanova", Email: "Testing@allo.com"},
		User{FistName: "Pride and Prejudice", LastName: "Jane Austen", Email: "Testing@allo.com"},
	}

	_, err = usersCollection.InsertMany(ctx, users)
	if err != nil {
		log.Fatalf("error creating default users %s", err.Error())
	}

	return usersCollection
}
