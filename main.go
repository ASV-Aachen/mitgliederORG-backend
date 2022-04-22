package main

import (
	"crypto/tls"
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

	"net/http"

	"github.com/ASV-Aachen/mitgliederDB-backend/keycloak"
)

// ------------------------------------------------------------------------------------------------------------
// GLOBAL VAR
var MariaDB *sql.DB
var PostgresDB *sql.DB
var err error

// MariaDB
var DB_USER string = os.Getenv("DB_USER")
var DB_PASSWORD string = os.Getenv("DB_PASSWORD")
var DB_NAME string = os.Getenv("DB_NAME")
var DB_URL string = os.Getenv("DB_URL")

// Postgres
var Postgreshost string = os.Getenv("Postgreshost")
var Postgresport int = 5432
var Postgresuser string = os.Getenv("Postgresuser")
var Postgrespassword string = os.Getenv("Postgrespassword")
var Postgresdbname string = os.Getenv("Postgresdbname")

// ------------------------------------------------------------------------------------------------------------
func main() {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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
		func(c *fiber.Ctx) error {
			token := c.Cookies("token")
			if token == "" {
				log.Fatalf("Token nicht gesendet")
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			// token = strings.Replace(token, "Bearer ", "", 1)

			ID, err := keycloak.Get_UserID(token)

			if err != nil {
				log.Default().Printf(err.Error())
				return c.SendStatus(401)
			}

			// Oauth login -> new Token
			newToken, err := keycloak.Get_AdminToken()

			if err != nil {
				log.Default().Printf(err.Error())
				return c.SendStatus(400)
			}

			userGroupes, err := keycloak.Get_UserGroups(newToken, ID)

			if err != nil {
				log.Default().Printf(err.Error())
				return c.SendStatus(416)
			}

			userGroups := [5]string{
				"Schriftwart",
				"Entwickler",
				"Admin",
			}

			if keycloak.Check_IsUserPartOfGroup(userGroups, userGroupes) {
				return c.Next()
			}

			return c.SendStatus(401)
		},
	)

	// Test handler
	api.Get("/health/", func(c *fiber.Ctx) error {
		return c.SendString("App running")
	})

	api.Get("/", func(c *fiber.Ctx) error {
		var erg_website [1000]website
		var erg_keycloak [1000]keycloak_user
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
				// &next.PROFILE_IMAGE,
				&next.STATUS)
			if err != nil {
				print("Error in WebsiteSQL: ")
				log.Default().Printf(err.Error())
			} else {
				erg_website[counter] = next
				counter++
			}
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
			var next keycloak_user
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

	api.Delete("/", func(c *fiber.Ctx) error {
		// Remove a User
		// TODO:
		return c.SendString("TODO: Remove a User")
	})

	// Sync all users every 3 hours

	api.Patch("/", func(c *fiber.Ctx) error {
		// Sync all Users
		url := "http://webpage:8080/api/sync"

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+c.Cookies("token"))
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			c.SendString(err.Error())
			panic(err)
		}

		if res.Status != "200 OK" {
			c.SendString(res.Status)
		}

		defer res.Body.Close()

		// Sync ArbeitsstundenDB
		// TODO:

		// Sync Bierkasse
		// TODO:

		return c.SendString("TODO: Sync all Users")
	})

	log.Fatal(app.Listen(":5000"))
}
