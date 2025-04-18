package types

type TaskScheduledToRegister struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Cron         string `json:"cron"`
	UrlToTrigger string `json:"url_to_trigger"`
}
