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

func main() {
	app.Build().Run()
}
