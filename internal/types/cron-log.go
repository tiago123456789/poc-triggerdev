package types

import "time"

type CronLog struct {
	Id         int       `json:"id"`
	ExternalId string    `json:"external_id"`
	Data       string    `json:"data"`
	CreatedAt  time.Time `json:"created_at"`
}
