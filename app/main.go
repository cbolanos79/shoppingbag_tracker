package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/api/idtoken"
)

type Login struct {
	Credential string `json:"credential"`
}

type UserProfile struct {
	Name       string `json:"name"`
	PictureUrl string `json:"picture_url"`
	AuthToken  string `json:"auth_token"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

var google_client_id string
var db *sql.DB
var jwt_signature string

func NewDB(adapter string, name string) (*sql.DB, error) {
	db, err := sql.Open(adapter, name)
	if err != nil {
		return nil, err
	}

	const create = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY,
		google_uid varchar(255)
	  );`

	if _, err := db.Exec(create); err != nil {
		return nil, err
	}

	return db, nil
}

// Check if given google id user exists in database
func checkIfGoogleUidExists(google_uid string) bool {
	row := db.QueryRow("SELECT id FROM users WHERE google_uid = ?", google_uid)

	var id int
	if err := row.Scan(&id); err != nil {
		return false
	}
	return true
}

// Receive credential for Google login and validate it agains Google API
// If credential is valid, extract name and profile picture url
// Else, returns an error
func login_google(c echo.Context) error {
	login := Login{}
	c.Bind(&login)

	// Return HTTP 422 if credential value is not set
	if len(login.Credential) == 0 {
		log.Println("Missing credential value")
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Missing credential value"})
	}

	// Validate credential with google client
	payload, err := idtoken.Validate(context.Background(), login.Credential, google_client_id)

	// Return HTTP 422 if there was any error
	if err != nil {
		log.Println("Error validating user in google: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error validating user"})
	}

	// Check if user exists
	if !checkIfGoogleUidExists(payload.Subject) {
		log.Printf("User %s not found", payload.Subject)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "User not found"})
	}

	// Check if token is expired
	if time.Now().Unix() > payload.Expires {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Expired credential"})
	}

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string)}

	// Return HTTP 200 if success
	return c.JSON(http.StatusOK, &userProfile)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	google_client_id = os.Getenv("GOOGLE_CLIENT_ID")
	if len(google_client_id) == 0 {
		panic("Empty value for GOOGLE_CLIENT_ID")
	}

	db_name := os.Getenv("DB_NAME")
	if len(db_name) == 0 {
		panic("Empty value for DB_NAME")
	}

	db, err = NewDB("sqlite3", db_name)
	if err != nil {
		log.Fatal(err)
		panic("Error opening database")
	}

	jwt_signature = os.Getenv("JWT_SIGNATURE")
	if len(jwt_signature) == 0 {
		panic("Missing jwt signature")
	}

	e := echo.New()
	e.Use(middleware.CORS())
	e.POST("/login/google", login_google)
	//e.POST("/ticket", create_ticket)
	e.Logger.Fatal(e.Start(":8000"))
}
