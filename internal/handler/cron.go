package handler

import (
	"log"
	"triggerdev-golang/internal/repository"
	"triggerdev-golang/internal/service"
	"triggerdev-golang/internal/types"

	"github.com/adhocore/gronx"
	"github.com/gofiber/fiber/v2"
)

type CronHandler struct {
	repository repository.ICronRepository
	service    *service.CronService
}

func CronHandlerNew(
	repository repository.ICronRepository,
	service *service.CronService,
) *CronHandler {
	return &CronHandler{
		repository: repository,
		service:    service,
	}
}

func (cH *CronHandler) GetCrons(c *fiber.Ctx) error {
	crons, err := cH.repository.GetCrons()
	if err != nil {
		log.Printf("%v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	return c.JSON(crons)
}

func (cH *CronHandler) RegisterCrons(c *fiber.Ctx) error {
	var data []types.TaskScheduledToRegister
	c.BodyParser(&data)

	mapExternalIds, err := cH.repository.GetAlreadyCreated()
	if err != nil {
		log.Printf("%v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	for _, cron := range data {

		if mapExternalIds[cron.Id] == true {
			err := cH.repository.Update(cron)
			if err != nil {
				log.Printf("failed to update user: %v", err)
				return c.Status(500).JSON(fiber.Map{
					"message": "Internal server error",
				})
			}
		} else {
			nextExecution, _ := gronx.NextTick(cron.Cron, false)
			err := cH.repository.Create(cron, nextExecution)
			if err != nil {
				log.Printf("failed to update user: %v", err)
				return c.Status(500).JSON(fiber.Map{
					"message": "Internal server error",
				})
			}
		}

	}
	return c.SendStatus(201)
}

func (cH *CronHandler) FinishedCronExecution(c *fiber.Ctx) error {
	err := cH.repository.SetCronAsNotExecuting(c.Params("externalId"))
	if err != nil {
		log.Printf("failed to update user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.SendStatus(204)
}

func (cH *CronHandler) TriggerCron(c *fiber.Ctx) error {
	err := cH.repository.SetCronAsNotExecuting(c.Params("externalId"))
	if err != nil {
		log.Printf("failed to update user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	cronToNotify, err := cH.repository.GetCronByIdEnabledAndNotExecuting(
		c.Params("externalId"),
	)
	if err != nil {
		log.Printf("failed to update user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	cH.service.TriggerCronOnThirdApplication(cronToNotify)
	return c.SendStatus(204)
}

func (cH *CronHandler) GetLogs(c *fiber.Ctx) error {
	cronLogs, err := cH.repository.GetLogs(c.Params("externalId"))
	if err != nil {
		log.Printf("failed to query: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(cronLogs)
}

func (cH *CronHandler) SaveLogs(c *fiber.Ctx) error {
	var logData map[string]interface{}
	c.BodyParser(&logData)

	err := cH.repository.SaveLog(logData)
	if err != nil {
		log.Printf("failed to insert: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.SendStatus(204)
}
