package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize template engine
	engine := html.New("./app", ".html")

	// Start a new fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Database Connection
	const DATABASE_URL string = "postgres://localhost:5432/time?sslmode=disable"

	dbpool, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	greeting := "hello mf"
	err = dbpool.QueryRow(context.Background(), "select 'hello'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	// API
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	app.Get("/practice", func(c *fiber.Ctx) error {
		return c.Render("message", fiber.Map{
			"Message": "HTMX Fucks",
		})
	})

	// form:"activity" specifices what field in
	// application/x-www-form-urlencoded
	type Timeblock struct {
		Id        int32     `db:"id"`
		Activity  string    `form:"activity" db:"activity"`
		TimeSpent int32     `form:"time" db:"time_spent"`
		CreatedAt time.Time `db:"id"`
		UpdatedAt time.Time `db:"id"`
		DeletedAt time.Time `db:"id"`
	}

	app.Get("/activity", func(c *fiber.Ctx) error {
		rows, _ := dbpool.Query(context.Background(), "SELECT * FROM timeblock WHERE deleted_at IS NULL")
		defer rows.Close()

		// Assuming rows is a pgx.Rows object obtained from a query
		var activities []Timeblock
		for rows.Next() {
			var tb Timeblock

			err := rows.Scan(&tb.Id, &tb.Activity, &tb.TimeSpent)
			if err != nil {
				fmt.Println(err)
			}
			// Now, tenant has been populated with values from the database row
			// Do something with the tenant data
			activities = append(activities, tb)
		}

		// Convert results slice to JSON
		// jsonResponse, err := json.Marshal(activities)
		// if err != nil {
		// 	log.Println("Error encoding JSON:", err)
		// 	return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON")
		// }

		// fmt.Println(string(jsonResponse))

		log.Println("Before marshalling JSON")
		jsonResponse, err := json.Marshal(activities)
		if err != nil {
			log.Println("Error encoding JSON:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON")
		}
		log.Println("After marshalling JSON")

		log.Println("JSON Response:", string(jsonResponse))

		return c.Render("index", jsonResponse)
		// return c.Render("activities", fiber.Map(rows.Values()))
	})

	app.Post("/activity", func(c *fiber.Ctx) error {
		tb := new(Timeblock)

		if err := c.BodyParser(tb); err != nil {
			return err
		}

		_, insertErr := dbpool.Exec(context.Background(), "INSERT INTO timeblock (activity, time_spent, created_at) VALUES ($1, $2, $3)", tb.Activity, tb.TimeSpent, time.Now())

		if insertErr != nil {
			fmt.Println(insertErr)
		}

		return c.Render("index", fiber.Map{
			"Title": "Activity Created",
		})
	})

	// Listen on PORT 300
	app.Listen(":3000")
}
