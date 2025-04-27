package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/RichCake/calc_api_go/agent/internal"
)

func init() {
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
	agent.RunAgent()
}