package main

import (
	"duels-api/app"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("no .env file found")
	}
}

// @title		Duels API
// @version		1.0
// @description	This is a swagger specification for a Duels back-end.
func main() {
	app.Build().Run()
}
