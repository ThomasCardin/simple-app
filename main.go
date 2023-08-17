package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Id        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	FirstName string             `bson:"FirstName,omitempty" json:"FirstName,omitempty"`
	LastName  string             `bson:"LastName,omitempty" json:"LastName,omitempty"`
	Email     string             `bson:"Email,omitempty" json:"Email,omitempty"`
}

type UserDTO struct {
	Id        string `bson:"_id" json:"id,omitempty"`
	FirstName string `bson:"FirstName,omitempty" json:"FirstName,omitempty"`
	LastName  string `bson:"LastName,omitempty" json:"LastName,omitempty"`
	Email     string `bson:"Email,omitempty" json:"Email,omitempty"`
}

func main() {
	mongoUser := os.Getenv(MONGO_USERNAME)
	mongoPassword := os.Getenv(MONGO_PASSWORD)
	mongoHost := os.Getenv(MONGO_HOST)
	serverPort := os.Getenv(SERVER_PORT)
	mongoDbUri := fmt.Sprintf("mongodb://%s:%s@%s:27017", mongoUser, mongoPassword, mongoHost)

	fmt.Printf("Connection string: %s", mongoDbUri)
	fmt.Printf("Server port: %s", serverPort)

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

		var users []UserDTO
		for cursor.Next(ctx) {
			var result User
			err := cursor.Decode(&result)
			if err != nil {
				fmt.Printf("error decoding documents: %s", err.Error())
			}

			users = append(users, UserDTO{
				Id:        result.Id.Hex(),
				FirstName: result.FirstName,
				LastName:  result.LastName,
				Email:     result.Email,
			})
		}

		c.HTML(200, "index.html", gin.H{
			"users": users,
		})
	})

	r.POST("/delete/:id", func(c *gin.Context) {
		id := c.Param("id")

		primitiveId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			fmt.Printf("Error converting the id to primitive.ObjectIDFromHex: %s", err.Error())
		}

		_, err = collection.DeleteOne(ctx, bson.M{"_id": primitiveId})
		if err != nil {
			fmt.Printf("Error deleting user with id: %s %s", id, err.Error())
		}

		c.Redirect(http.StatusMovedPermanently, "/")
	})

	r.POST("/add", func(c *gin.Context) {
		firstName := c.PostForm("FirstName")
		lastName := c.PostForm("LastName")
		email := c.PostForm("Email")

		user := User{
			Id:        primitive.NewObjectID(),
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

	err := r.Run(fmt.Sprintf(":%s", serverPort))
	if err != nil {
		fmt.Printf("Error starting the GIN server: %s", err.Error())
	}
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
		User{Id: primitive.NewObjectID(), FirstName: "Cat's Cradle", LastName: "Kurt Vonnegut Jr.", Email: "Testing@allo.com"},
		User{Id: primitive.NewObjectID(), FirstName: "Robert", LastName: "Bob", Email: "foo@allo.com"},
		User{Id: primitive.NewObjectID(), FirstName: "George", LastName: "Annie", Email: "bar@allo.com"},
	}

	_, err = usersCollection.InsertMany(ctx, users)
	if err != nil {
		fmt.Printf("error creating default users %s", err.Error())
	}

	return usersCollection
}
