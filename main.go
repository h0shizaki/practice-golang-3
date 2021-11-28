package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname" bson:"firstname, omitempty"`
	Lastname  string             `json:"lastname" bson:"lastname, omitempty"`
}

var client *mongo.Client

func sayHello(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Hello GO"))
}

func createPerson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)

	collection := client.Database("mydb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.InsertOne(ctx, person)

	if err != nil {
		log.Println("There is an Error ", err)
	}

	json.NewEncoder(res).Encode(result)

}

func getAllPerson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	var people []Person

	collection := client.Database("mydb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		log.Print("ERROR")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}

	if err := cursor.Err(); err != nil {
		log.Print("ERROR")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(res).Encode(people)
}

func getPerson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Conetent-Type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	// log.Print(id)
	var person Person

	collection := client.Database("mydb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&person)

	if err != nil {
		log.Print("ERROR")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(res).Encode(person)
}

func main() {
	log.Println("Start server")

	router := mux.NewRouter()

	//Conecct to Database
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	//COR
	header := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origin := handlers.AllowedOrigins([]string{"*"})

	//Route
	router.HandleFunc("/", sayHello).Methods("GET")
	router.HandleFunc("/people", getAllPerson).Methods("GET")
	router.HandleFunc("/person/{id}", getPerson).Methods("GET")
	router.HandleFunc("/person", createPerson).Methods("POST")

	//Run
	var port string = ":3030"
	log.Println("Server is running on port", port)
	http.ListenAndServe(port, handlers.CORS(header, methods, origin)(router))
}
