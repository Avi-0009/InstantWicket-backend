package main

import (
	"log"

	"github.com/Avi-0009/InstantWicket-backend/config"
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/server"
)

func main() {
	config.LoadConfig()

	err := database.ConnectandMigrate(
		config.GetEnv("DB_HOST", "-"),
		config.GetEnv("DB_PORT", "-"),
		config.GetEnv("DB_NAME", "-"),
		config.GetEnv("DB_USER", "-"),
		config.GetEnv("DB_PASSWORD", "password"),
		database.SSLMode(config.GetEnv("DB_SSLMode", "disable")),
	)
	if err != nil {
		log.Fatal(err)
	}

	r := server.NewServer()
	port := config.GetEnv("PORT", "8080")
	r.Run(":" + port)
}
