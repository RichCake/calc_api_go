package main

import "github.com/RichCake/calc_api_go/internal/application"

func main() {
	app := application.New()
	app.RunServer()
}