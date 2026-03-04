package main

import "github.com/Final-Year-Project-G22/backend/core/internal/bootstrap"

func main() {
	app := bootstrap.NewApp()
	app.Run()
}
