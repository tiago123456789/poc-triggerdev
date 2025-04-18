## ABOUT

The project has goal to challlenge myself to create a solution like trigger.dev where has
a sdk and allow you to setup functions to execute based cronjob expression without using cronjob libraries.

But my solution is focus in application running on serverless container such as: **cloud run** where I the application
doesn't have interaction the application is stopped, but the problem is the cronjob routines can execute only if the application is running.

In my project I resolved it doing the following strategy:

- Your application will expose a URL.
- The scheduler will notify the URL about the cronjob needs to be executed, when the scheduler send a request the serverless container solution will reactivate the application to process the request.
- The SDK I created will get the data received from request and finally execute the function.

## TECHNOLOGIES:

### API

- Golang
- Fiber
- TursoDB(database)

### SDK

- Golang

## INSTRUCTIONS TO RUN THE API

- Clone
- Create file .env with the envs:

```txt

TURSO_DATABASE_URL="" // Create account on Turso DB to get that data
TURSO_AUTH_TOKEN=""  // Create account on Turso DB to get that data
```

- Execute command **go run cmd/api/main.go** to start the api. PS: if you starting the application first time, will download the packages and after will create the tables in TursoDB.
- Execute command **go run cmd/scheduler/main.go** to start the scheduler responsible to notify cronjob if needs to trigger.

## INSTRUCTIONS TO USE THE SDK ON YOUR API

- Access your project
- Execute the command **go get github.com/tiago123456789/poc-triggerdev-sdk@v1.0.0** to download the sdk to register the cronjob tasks.
- Set the following envs:

```txt
REMOTE_TRIGGER_ENDPOINT="address_from_api_path_cron" // For example: http://localhost:3000/crons
REMOTE_TRIGGER_LOGGERS_ENDPOINT="address_from_api_path_cron" // For example: http://localhost:3000/crons-logs
URL_TO_TRIGGER="address_from_your_application_scheduler_notify" // For example: http://localhost:3002/crons
```

- Example using the SDK:

```golang

package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/tiago123456789/poc-triggerdev-sdk/task"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	taskScheduled := task.Init()

	taskScheduled.Add(task.TaskScheduled{
		Id:   "hellov2",
		Name: "Say hello",
		Cron: "*/1 * * * *",
		Action: func(message map[string]interface{}, logger *slog.Logger) error {
			logger.Info("Executing code")
			fmt.Println("Hi my friend v2!!!")
			return nil
		},
	})

	taskScheduled.Add(task.TaskScheduled{
		Id:   "hi-testv2",
		Name: "hi-test",
		Cron: "*/1 * * * *",
		Action: func(message map[string]interface{}, logger *slog.Logger) error {
			logger.Info("Executing code")
			fmt.Println("hi-test v2")
			return nil
		},
	})

	taskScheduled.Start()

	app := fiber.New()

	app.Post("/crons", func(c *fiber.Ctx) error {
		var data task.TaskScheduledToRegister
		c.BodyParser(&data)

		go taskScheduled.Execute(data.Id, map[string]interface{}{})
		return c.SendString("OK")
	})

	app.Listen(":3002")
}
```

- Now you need to start your application, so the sdk will save the cronjob tasks data in api and the scheduler will start to notify the cronjobs needs to be executed.

## ARCHITECTURE

![Arquitetura](./architecture.png)
