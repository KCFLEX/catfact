package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// we make a server
type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server { // constructor function
	return &Server{
		client: c,
	}
}

// now we create api handlers
// this handler get access to the collection "facts" in the "catfacts" database and retrieves all the data inthat collection in json form
func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("catfact").Collection("facts") // this line creates a variable coll that represents a MongoDB collection named "facts" within the "catfact" database.
	// It uses the s.client field, which is expected to be a MongoDB client, to access the database and collection.

	query := bson.M{} // This line defines a query using a BSON Map (bson.M) to specify that you want to retrieve all documents in the collection.
	//An empty map means no specific filtering conditions.
	cursor, err := coll.Find(context.TODO(), query) // This line queries the MongoDB collection using the Find method. It takes a context,
	//which is set to context.TODO() for an empty context, and the query defined in the previous line.
	//and the result is stored in the cursor variable
	if err != nil {
		log.Fatal(err)
	}
	results := []bson.M{} // slice of bson.M
	// checking errors in conversion
	if err = cursor.All(context.TODO(), &results); err != nil { //This line retrieves all documents from the cursor and stores them in the results slice.
		//It also checks for any errors in the conversion process.
		panic(err)
	}
	w.WriteHeader(http.StatusOK)                       // This line sets the HTTP response status code to 200 OK, indicating that the request was successful.
	w.Header().Add("Content-Type", "application/json") // This line adds an HTTP response header to specify that the response is in JSON format.
	json.NewEncoder(w).Encode(results)                 // This line encodes the results slice as JSON and sends it as the response body to the client via the w HTTP response writer.

}

// we are going to make a simple catfact worker that is going to spin up asynchronously along side our jsonapi and basically gonna fetch facts about cats
type CatfactWorker struct {
	client *mongo.Client // It's designed to hold a reference to a MongoDB client instance, allowing other methods and functions to work with MongoDB.
}

// The primary purpose of constructor functions is to encapsulate the details of struct initialization and provide a clean and consistent way to create instances
func NewCatFactWorker(c *mongo.Client) *CatfactWorker { // constructor function: It is a regular function that returns a newly created instance of the struct.
	return &CatfactWorker{
		client: c,
	}
}

// func (cfw *CatfactWorker) start() error is a method intended to start the worker to periodically fetch and insert cat facts into the MongoDB databaseusing a provided MongoDB client.
func (cfw *CatfactWorker) start() error {
	coll := cfw.client.Database("catfact").Collection("facts") //This line initializes a coll variable by accessing the "facts" collection in the "catfact" database using the client stored in the CatfactWorker. It prepares to interact with this collection
	ticker := time.NewTicker(2 * time.Minute)                  //This line creates a ticker using the time.NewTicker function, which will tick every 2 minutes. This ticker will be used to control the periodic fetching and insertion of cat facts.

	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M                                                  // map[string]any | This line declares a variable catFact of type bson.M, which is a BSON map used to store the cat fact data.
		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil { //to decode the JSON response from the API into the catFact variable. If there's an error during decoding, the method returns that error.
			return err
		}
		// store catfact in mongodb
		_, err = coll.InsertOne(context.TODO(), catFact)
		if err != nil {
			return err
		}
		<-ticker.C // This code blocks until the ticker ticks (every 2 minutes). After that, the loop continues, fetching another cat fact and inserting it into the database.
	}
}
func main() {
	// this code establishes a connection to a MongoDB database running on localhost at port 27017 using the Go MongoDB driver. If an error occurs during the connection attempt, it will panic and print the error message. Otherwise, it will store the MongoDB client in the client variable for further use.
	URI := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(URI))
	if err != nil {
		panic(err)
	}

	worker := NewCatFactWorker(client) // This line creates a CatFactWorker instance by calling the NewCatFactWorker constructor function and passing the MongoDB client (client) as a parameter. This worker is designed to fetch cat facts and insert them into the MongoDB database.
	//This code starts the worker asynchronously using the go keyword. It spawns a new goroutine to execute the start method of the worker.
	//The purpose of making it asynchronous is to ensure that the worker can continuously fetch and insert cat facts without blocking the main program, which allows other tasks to be performed simultaneously.
	go worker.start()

	server := NewServer(client)                         // boot up server
	http.HandleFunc("/facts", server.handleGetAllFacts) //This line registers a handler function (server.handleGetAllFacts) to respond to HTTP requests at the "/facts" endpoint. When a client makes an HTTP request to this endpoint, the handleGetAllFacts function will be invoked.
	http.ListenAndServe(":3000", nil)                   // This code starts an HTTP server that listens on port 3000 and serves incoming HTTP requests.
}

// in summary, this code connects to a MongoDB database,
//starts a worker to fetch and insert cat facts asynchronously,
//and spins up an HTTP server to serve API requests related to cat facts.
//It uses goroutines to handle concurrent execution and panics
//if an error occurs during the database connection.
