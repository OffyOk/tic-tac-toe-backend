package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var gameCollection *mongo.Collection

type Game struct {
	ID      string   `json:"id,omitempty" bson:"_id,omitempty"`
	Squares []string `json:"squares" bson:"squares"`
	IsXNext bool     `json:"isXNext" bson:"isXNext"`
}

func connectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	gameCollection = client.Database("xo_game").Collection("games")
}

func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	var games []Game
	cursor, err := gameCollection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var game Game
		cursor.Decode(&game)
		games = append(games, game)
	}
	json.NewEncoder(w).Encode(games)
}

func createGameHandler(w http.ResponseWriter, r *http.Request) {
	var game Game
	json.NewDecoder(r.Body).Decode(&game)
	result, err := gameCollection.InsertOne(context.Background(), game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game.ID = result.InsertedID.(string)
	json.NewEncoder(w).Encode(game)
}

func main() {
	connectDB()
	r := mux.NewRouter()
	r.HandleFunc("/api/games", getGamesHandler).Methods("GET")
	r.HandleFunc("/api/games", createGameHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
