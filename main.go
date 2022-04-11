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

	"bytes"
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

func sendToWeb(Mail, first_name, last_name, Entrydate, status, token string) bool {
	path := "webpage/api/addMember"

	values := map[string]string{
		"mail":       Mail,
		"first_name": first_name,
		"last_name":  last_name,
		"entrydate":  Entrydate,
		"status":     status,
	}

	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(json_data))

	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer resp.Body.Close()

	return true
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
				"Bierwart",
				"Entwickler",
				"Admin",
			}

			if keycloak.Check_IsUserPartOfGroup(userGroups, userGroupes) {
				return c.Next()
			}

			return c.SendStatus(417)
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
	
	api.Post("/", func(c *fiber.Ctx) error {
		// Add new User

		payload := struct {
			Mail       string `json:"mail"`
			First_name string `json:"first_name"`
			Last_name  string `json:"last_name"`
			EntryDate  string `json:"entryDate"`
			Status     string `json:"status"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		token := c.Cookies("keycloakToken")

		success := sendToWeb(payload.Mail, payload.First_name, payload.Last_name, payload.EntryDate, payload.Status, token)

		// TODO: in ArbeitsstundenDB eintragen

		if success {
			return c.Status(fiber.StatusOK).SendString("Nutzer angelegt")
		} else {
			return c.Status(fiber.StatusBadRequest).SendString("Es konnte kein Nutzer angelegt werden")
		}
	})

	api.Delete("/", func(c *fiber.Ctx) error {
		// Remove a User
		return c.SendString("Remove a User")
	})
	api.Patch("/", func(c *fiber.Ctx) error {
		// update Users in ArbeitsstundenDB

		return c.SendString("Change a User")
	})

	log.Fatal(app.Listen(":5000"))
}
