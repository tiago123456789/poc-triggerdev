package main

import (
	"log"
	"triggerdev-golang/internal/config"
	"triggerdev-golang/internal/handler"
	"triggerdev-golang/internal/repository"
	"triggerdev-golang/internal/service"

	"github.com/gofiber/fiber/v2"
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

	app := fiber.New()

	cronRepository := repository.CronRepositoryNew(db)
	cronService := service.CronServiceNew(cronRepository)
	cronHandler := handler.CronHandlerNew(cronRepository, cronService)

	app.Post("/crons-logs", cronHandler.SaveLogs)
	app.Get("/crons/:externalId/logs", cronHandler.GetLogs)
	app.Get("trigger-cronjobs/:externalId", cronHandler.TriggerCron)
	app.Post("/crons-finished-execution/:externalId", cronHandler.FinishedCronExecution)
	app.Post("/crons", cronHandler.RegisterCrons)
	app.Get("/crons", cronHandler.GetCrons)

	app.Listen(":3000")
}
