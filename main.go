package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/helmet/v2"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"net/http"

	"github.com/ASV-Aachen/mitgliederDB-backend/database"
	"github.com/ASV-Aachen/mitgliederDB-backend/keycloak"
	"github.com/google/uuid"
)

// ------------------------------------------------------------------------------------------------------------
// GLOBAL VAR
var MariaDB *sql.DB
var PostgresDB *sql.DB
var err error

func getUserData(MariaDB *sql.DB, PostgresDB *sql.DB) jsonStruct {
	var erg jsonStruct = jsonStruct{}

	var erg_website [1000]website
	var erg_keycloak [1000]keycloak_user
	var erg_arbeit [1000]arbeitsstunden

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

		}
		erg_keycloak[counter] = next
		counter++
	}
	keycloakRows.Close()
	erg.Keycloak = erg_keycloak[:counter]

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

	return erg
}

func SendSyncToWebsite(c *fiber.Ctx) {
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
}

func getStatus(s string) string {
	switch s {
	case "1":
		return "PROSPECT"
	case "2":
		return "ACTIVE"
	case "3":
		return "INACTIVE"
	default:
		return "OLD_MAN"
	}
}

// ------------------------------------------------------------------------------------------------------------
func main() {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Setup DATABASEs
	// MariaDB := database.SetUpMariaDB_admin()
	MariaDB = database.SetUpMariaDB()
	PostgresDB := database.SetUpPostgres()

	// Setup Views
	database.ExecuteFile(MariaDB, "MariaDB_createView_Mitglieder.sql")
	database.ExecuteFile(MariaDB, "MariaDB_createView_Keycloak.sql")

	database.ExecuteFile(PostgresDB, "PostgresSQL_createView.sql")

	defer MariaDB.Close()
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
		var erg jsonStruct = getUserData(MariaDB, PostgresDB)

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return json.NewEncoder(c.Response().BodyWriter()).Encode(erg)
	})

	api.Delete("/", func(c *fiber.Ctx) error {
		// Remove a User
		// TODO:
		return c.SendString("TODO: Remove a User")
	})

	api.Patch("/", func(c *fiber.Ctx) error {
		// Sync all Users of Webpage
		SendSyncToWebsite(c)

		// Sync ArbeitsstundenDB
		// DONE:

		var UserData jsonStruct = getUserData(MariaDB, PostgresDB)
		checker := false

		var finalUsers_Keycloak [1000]arbeitstundenInput
		counter := 0

		for _, KeycloakUser := range UserData.Keycloak {
			for _, arbeitsUser := range UserData.Arbeitsstunden {
				// Wer ist nicht da?
				// Ist die Person Aktiv?
				if KeycloakUser.EMAIL == arbeitsUser.EMAIL {
					checker = true
					break
				}
			}
			if checker == false {
				finalUsers_Keycloak[counter] = arbeitstundenInput{
					first_name: KeycloakUser.FIRST_NAME,
					last_name:  KeycloakUser.LAST_NAME,
					email:      KeycloakUser.EMAIL,
					user_id:    uuid.New(),
					member_id:  uuid.New(),
					reduction:  0,
					role:       "USER",
					password:   "$2a$10$ljHyydV.cFZAZJCsWWbmFOuvjiKnj1lOw.3ynYkl6GOTOF8OTxoaG",
				}
				counter++
			}
			checker = false
		}

		for _, websiteUser := range UserData.Website {
			for _, FutureArbeitsUser := range finalUsers_Keycloak {
				if websiteUser.EMAIL == FutureArbeitsUser.email {
					status := getStatus(websiteUser.STATUS)
					FutureArbeitsUser.status = status
					break
				}
			}
		}

		t := time.Now()
		year := t.Year()

		print(counter)
		print("\n")

		// create new member
		for _, newUser := range finalUsers_Keycloak[:counter] {
			query := "INSERT INTO member(id,user_id,first_name,last_name) VALUES ($1, $2, $3, $4);"
			_, err := PostgresDB.Exec(query, newUser.member_id, newUser.user_id, newUser.first_name, newUser.last_name)
			if err != nil {
				print(err.Error() + "    ")
				print("Member Error: " + newUser.first_name + " " + newUser.last_name + "\n")
				continue
			}

			query = "INSERT INTO user_(id,member_id,email,password,role) VALUES ($1, $2, $3, $4, $5);"
			_, err = PostgresDB.Exec(query, newUser.user_id, newUser.member_id, newUser.email, newUser.password, newUser.role)
			if err != nil {
				print("User Error: " + newUser.first_name + " " + newUser.last_name + " " + newUser.user_id.String() + " " + newUser.member_id.String() + "\n")
				print(err.Error())
				continue
			}

			query = "INSERT INTO reduction(id,season,member_id,status,reduction) VALUES ($1, $2, $3, $4, $5);"
			_, err = PostgresDB.Exec(query, newUser.user_id, year, newUser.member_id, newUser.status, newUser.reduction)
			if err != nil {
				print("Reduction Error: " + newUser.first_name + " " + newUser.last_name + newUser.user_id.String() + newUser.member_id.String() + "\n")
				print(err.Error())
				continue
			}
		}

		// Sync Bierkasse
		// TODO:

		return c.SendString("Synced all Users")
	})

	log.Fatal(app.Listen(":5000"))
}
