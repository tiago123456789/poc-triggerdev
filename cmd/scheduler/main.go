package main

import (
	"log"
	"triggerdev-golang/internal/config"
	"triggerdev-golang/internal/repository"
	"triggerdev-golang/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.InitDB()
	defer db.Close()

	cronRepository := repository.CronRepositoryNew(db)
	cronService := service.CronServiceNew(cronRepository)
	cronService.StartScheduler()
}
