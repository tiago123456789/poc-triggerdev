package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"triggerdev-golang/internal/repository"
	"triggerdev-golang/internal/types"

	"github.com/adhocore/gronx"
)

type CronService struct {
	repository repository.ICronRepository
}

func CronServiceNew(
	repository repository.ICronRepository,
) *CronService {
	return &CronService{
		repository: repository,
	}
}

func (c *CronService) notifyInParallel(itemsToTrigger []types.TaskScheduledToRegister) {
	var wg sync.WaitGroup

	for _, item := range itemsToTrigger {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.TriggerCronOnThirdApplication(item)
			nextExecution, _ := gronx.NextTick(item.Cron, false)
			err := c.repository.SetNextExecution(item, nextExecution)
			if err != nil {
				log.Printf("failed to update user: %v", err)
			}
		}()
	}

	wg.Wait()
}

func (c *CronService) StartScheduler() {
	log.Printf("Starting scheduler")

	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			itemsToTrigger, err := c.repository.GetCronsToTrigger()
			if err != nil {
				log.Printf("failed to get crons to trigger: %v", err)
			}

			log.Printf("%d total of crons to trigger now", len(itemsToTrigger))
			c.notifyInParallel(itemsToTrigger)
			log.Printf("Notified all %d crons", len(itemsToTrigger))

		}
	}
}

func (c *CronService) TriggerCronOnThirdApplication(item types.TaskScheduledToRegister) {
	jsonData, err := json.Marshal(item)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	req, err := http.NewRequest("POST", item.UrlToTrigger, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	defer resp.Body.Close()
}
