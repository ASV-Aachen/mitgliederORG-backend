package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
    app := fiber.New()
    app.Use(cors.New())

    api := app.Group("/MitgliederDB/api")

    // Test handler
    api.Get("/health/", func(c *fiber.Ctx) error {
        return c.SendString("App running")
    })

	api.Get("/", func(c *fiber.Ctx) error {
		// Get all Useres
		return c.SendString("Get Users")
	})
	api.Post("/", func(c *fiber.Ctx) error {
		// Add new User
		return c.SendString("Add new User")	
	})
	api.Delete("/", func(c *fiber.Ctx) error {
		// Remove a User
		return c.SendString("Add new User")	
	})
	api.Patch("/", func(c *fiber.Ctx) error {
		// Change a User
		return c.SendString("Add new User")	
	})


    log.Fatal(app.Listen(":5000"))
}
