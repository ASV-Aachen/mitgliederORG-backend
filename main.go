package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/helmet/v2"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// ------------------------------------------------------------------------------------------------------------
// GLOBAL VAR
var MariaDB *sql.DB
var PostgresDB *sql.DB
var err error

var DB_USER string = os.Getenv("DB_USER")
var DB_PASSWORD string = os.Getenv("DB_PASSWORD")
var DB_NAME string = os.Getenv("DB_NAME")
var DB_URL string = os.Getenv("DB_URL")

var Postgreshost string = os.Getenv("Postgreshost")
var Postgresport int = 5432
var Postgresuser string = os.Getenv("Postgresuser")
var Postgrespassword string = os.Getenv("Postgrespassword")
var Postgresdbname string = os.Getenv("Postgresdbname")

// ------------------------------------------------------------------------------------------------------------
// SQL Structs
type jsonStruct struct {
	Website        []website        `json:"WEBSITE"`
	Keycloak       []keycloak       `json:"KEYCLOAK"`
	Arbeitsstunden []arbeitsstunden `json:"ARBEITSSTUNDEN"`
}

type website struct {
	ID            string `json:"id"`
	USERNAME      string `json:"username"`
	FIRST_NAME    string `json:"first_name"`
	LAST_NAME     string `json:"last_name"`
	EMAIL         string `json:"email"`
	IS_ACTIVE     bool   `json:"is_active"`
	PROFILE_IMAGE string `json:"profile_image"`
	STATUS        string `json:"status"`
}

type keycloak struct {
	ID             string `json:"id"`
	EMAIL          string `json:"Email"`
	EMAIL_VERIFIED []byte `json:"Email_verified"`
	ENABLED        []byte `json:"enabled"`
	FIRST_NAME     string `json:"first_name"`
	LAST_NAME      string `json:"last_name"`
	USERNAME       string `json:"username"`
}

type arbeitsstunden struct {
	FIRST_NAME string `json:"first_name"`
	LAST_NAME  string `json:"last_name"`
	EMAIL      string `json:"email"`
}

// ------------------------------------------------------------------------------------------------------------
func main() {
	// Setup DATABASE
	MariaDB, err = sql.Open("mysql", DB_USER+":"+DB_PASSWORD+"@tcp("+DB_URL+":3306"+")/"+DB_NAME)
	if err != nil {
		panic(err.Error())
	}
	defer MariaDB.Close()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", Postgreshost, Postgresport, Postgresuser, Postgrespassword, Postgresdbname)
	// PostgresDB, err := sql.Open("postgres", Postgresuser+":"+Postgrespassword+"@tcp("+Postgreshost+":5432"+")/"+Postgresdbname)
	PostgresDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer PostgresDB.Close()

	// Setup
	app := fiber.New()
	app.Use(cors.New())

	api := app.Group("/mitgliederDB/api")

	// Use middlewares for each route
	app.Use(
		// add Helmet middleware
		helmet.New(),

		// Check if User is Part of Keyclaok and is logged in
		// TODO: to implement
		func(c *fiber.Ctx) error {
			return c.Next()
		},
	)

	// Test handler
	api.Get("/health/", func(c *fiber.Ctx) error {
		return c.SendString("App running")
	})

	api.Get("/", func(c *fiber.Ctx) error {
		var erg_website [1000]website
		var erg_keycloak [1000]keycloak
		var erg_arbeit [1000]arbeitsstunden

		var erg jsonStruct = jsonStruct{}

		// Website in Mariadb SELECT * FROM Website;
		websiteRows, err := MariaDB.Query("SELECT * FROM Website;")
		if err != nil {
			panic(err)
		}

		counter := 0
		for websiteRows.Next() {
			var next website
			err := websiteRows.Scan(
				&next.ID,
				&next.USERNAME,
				&next.FIRST_NAME,
				&next.LAST_NAME,
				&next.EMAIL,
				&next.IS_ACTIVE,
				&next.PROFILE_IMAGE,
				&next.STATUS)
			if err != nil {
				print("Error in WebsiteSQL: ")
				log.Fatal(err)
			}

			erg_website[counter] = next
			counter++
		}
		websiteRows.Close()
		erg.Website = erg_website[:counter]

		// Keycloak in MariaDB SELECT * FROM Keycloak;
		keycloakRows, err := MariaDB.Query("SELECT * FROM Keycloak;")
		if err != nil {
			panic(err)
		}

		counter = 0
		for keycloakRows.Next() {
			var next keycloak
			err := keycloakRows.Scan(
				&next.ID,
				&next.EMAIL,
				&next.EMAIL_VERIFIED,
				&next.ENABLED,
				&next.FIRST_NAME,
				&next.LAST_NAME,
				&next.USERNAME,
			)
			if err != nil {
				print("Error in KeycloakDB: ")
				print(err.Error())
				print("\n")
				// log.Fatal(err)
			}
			erg_keycloak[counter] = next
			counter++
		}
		keycloakRows.Close()
		erg.Keycloak = erg_keycloak[:counter]

		// ArbeitsstundenMitglieder in Postgres SELECT * FROM arbeitsstundenmitglieder;
		arbeitsstundenRows, err := PostgresDB.Query("SELECT * FROM arbeitsstundenmitglieder")
		if err != nil {
			print("\n\n ---------------- \n\n")
			print("Error in Connecting to POSTGRESDB \n")
			panic(err)
		}

		counter = 0
		for arbeitsstundenRows.Next() {
			var next arbeitsstunden
			err := arbeitsstundenRows.Scan(
				&next.FIRST_NAME,
				&next.LAST_NAME,
				&next.EMAIL,
			)
			if err != nil {
				print("Error in ArbeitsstundenDB: ")
				log.Fatal(err)
			}
			erg_arbeit[counter] = next
			counter++
		}
		arbeitsstundenRows.Close()
		erg.Arbeitsstunden = erg_arbeit[:counter]

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return json.NewEncoder(c.Response().BodyWriter()).Encode(erg)
	})
	api.Post("/", func(c *fiber.Ctx) error {
		// Add new User
		return c.SendString("Add new User")
	})
	api.Delete("/", func(c *fiber.Ctx) error {
		// Remove a User
		return c.SendString("Remove a User")
	})
	api.Patch("/", func(c *fiber.Ctx) error {
		// Change a User
		return c.SendString("Change a User")
	})

	log.Fatal(app.Listen(":5000"))
}
