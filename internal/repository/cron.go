package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
	"triggerdev-golang/internal/types"
)

type ICronRepository interface {
	GetLogs(externalId string) ([]types.CronLog, error)
	SaveLog(logData map[string]interface{}) error
	SetCronAsNotExecuting(externalId string) error
	GetCronByIdEnabledAndNotExecuting(
		externalId string,
	) (types.TaskScheduledToRegister, error)
	GetAlreadyCreated() (map[string]bool, error)
	Create(
		cron types.TaskScheduledToRegister, nextExecution time.Time) error
	Update(
		cron types.TaskScheduledToRegister) error
	GetCronsToTrigger() ([]types.TaskScheduledToRegister, error)
	SetNextExecution(
		item types.TaskScheduledToRegister, nextExecution time.Time) error
	GetCrons() ([]types.TaskScheduledToRegister, error)
}

type CronRepository struct {
	db *sql.DB
}

func CronRepositoryNew(db *sql.DB) *CronRepository {
	return &CronRepository{
		db: db,
	}
}

func (c *CronRepository) GetCrons() ([]types.TaskScheduledToRegister, error) {
	rows, err := c.db.QueryContext(context.Background(), `
		select name, external_id, expression, url_to_trigger from cronjobs
		order by created_at asc
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	itemsToTrigger := []types.TaskScheduledToRegister{}
	for rows.Next() {
		var item types.TaskScheduledToRegister
		if err := rows.Scan(&item.Name, &item.Id, &item.Cron, &item.UrlToTrigger); err != nil {
			return nil, err
		}

		itemsToTrigger = append(itemsToTrigger, item)
	}

	return itemsToTrigger, nil
}

func (c *CronRepository) GetCronsToTrigger() ([]types.TaskScheduledToRegister, error) {
	rows, err := c.db.QueryContext(context.Background(), `
	select external_id, expression, url_to_trigger from cronjobs where 
	current_timestamp >= next_execution and enabled = true and is_executing = false
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	itemsToTrigger := []types.TaskScheduledToRegister{}
	for rows.Next() {
		var item types.TaskScheduledToRegister
		if err := rows.Scan(&item.Id, &item.Cron, &item.UrlToTrigger); err != nil {
			return nil, err
		}

		itemsToTrigger = append(itemsToTrigger, item)
	}

	return itemsToTrigger, nil
}

func (c *CronRepository) SetNextExecution(
	item types.TaskScheduledToRegister, nextExecution time.Time) error {
	_, err := c.db.ExecContext(context.Background(), `
							UPDATE cronjobs SET next_execution = ?, is_executing = true
							WHERE external_id = ?
						`, nextExecution.UTC(), item.Id)
	if err != nil {
		return err
	}

	return nil
}

func (c *CronRepository) Update(
	cron types.TaskScheduledToRegister) error {
	_, err := c.db.ExecContext(context.Background(), `
		UPDATE cronjobs SET name = ?, expression = ?, url_to_trigger = ?
		WHERE external_id = ?
	`, cron.Name, cron.Cron, cron.UrlToTrigger, cron.Id)

	if err != nil {
		return err
	}

	return nil
}

func (c *CronRepository) Create(
	cron types.TaskScheduledToRegister, nextExecution time.Time) error {
	_, err := c.db.ExecContext(context.Background(), `
					INSERT INTO cronjobs (external_id, name, expression, next_execution, url_to_trigger) 
					VALUES (?, ?, ?, ?, ?)
				`, cron.Id, cron.Name, cron.Cron, nextExecution.UTC(), cron.UrlToTrigger)
	if err != nil {
		return err
	}

	return nil
}

func (c *CronRepository) GetAlreadyCreated() (map[string]bool, error) {
	rows, err := c.db.QueryContext(
		context.Background(),
		"SELECT external_id FROM cronjobs",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapExternalIds := map[string]bool{}
	for rows.Next() {
		var externalId string
		if err := rows.Scan(&externalId); err != nil {
			return nil, err
		}
		mapExternalIds[externalId] = true
	}

	return mapExternalIds, nil
}

func (c *CronRepository) GetCronByIdEnabledAndNotExecuting(
	externalId string,
) (types.TaskScheduledToRegister, error) {
	rows, err := c.db.QueryContext(context.Background(), `
				select external_id, expression, url_to_trigger from cronjobs where 
					enabled = true and is_executing = false and external_id = ? LIMIT 1
			`, externalId)
	if err != nil {
		return types.TaskScheduledToRegister{}, err
	}

	defer rows.Close()
	var item types.TaskScheduledToRegister
	for rows.Next() {
		if err := rows.Scan(&item.Id, &item.Cron, &item.UrlToTrigger); err != nil {
			return types.TaskScheduledToRegister{}, err
		}

	}

	return item, nil
}

func (c *CronRepository) SetCronAsNotExecuting(externalId string) error {
	_, err := c.db.ExecContext(context.Background(), `
			UPDATE cronjobs SET is_executing = false
			WHERE external_id = ?
		`, externalId)
	if err != nil {
		return err
	}

	return nil
}

func (c *CronRepository) SaveLog(logData map[string]interface{}) error {
	logDataToSave, err := json.Marshal(logData)

	_, err = c.db.ExecContext(context.Background(), `
				INSERT INTO cronjobs_logs (external_id, data) 
				VALUES (?, ?)
			`, logData["id"], logDataToSave)

	if err != nil {
		return err
	}

	return nil
}

func (c *CronRepository) GetLogs(externalId string) ([]types.CronLog, error) {
	rows, err := c.db.QueryContext(context.Background(), `
		SELECT id, external_id, data, created_at FROM cronjobs_logs
		where external_id = ? order by created_at desc LIMIT 20
	`, externalId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cronLogs := []types.CronLog{}
	for rows.Next() {
		var cronLog types.CronLog
		if err := rows.Scan(
			&cronLog.Id, &cronLog.ExternalId,
			&cronLog.Data, &cronLog.CreatedAt,
		); err != nil {
			log.Fatal(err)
		}

		cronLogs = append(cronLogs, cronLog)
	}

	return cronLogs, nil
}
