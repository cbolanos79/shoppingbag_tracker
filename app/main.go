package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "google.golang.org/api/idtoken"
    "github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

type Login struct {
	Credential string `json:"credential"`
}

type UserProfile struct {
    Name string `json:"name"`
    Picture_url string `json:"picture_url"`
}

var google_client_id string

// Receive credential for Google login and validate it agains Google API
// If credential is valid, extract name and profile picture url
// Else, returns an error
func login_google(c echo.Context) error {
	login := Login{}
	c.Bind(&login)

	// Return HTTP 422 if credential value is not set
	if len(login.Credential) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Missing credential value"})
	}

	// Validate credential with google client
	payload, err := idtoken.Validate(context.Background(), login.Credential, google_client_id)

	// Return HTTP 422 if there was any error
	if err != nil {
        log.Println("Error validating user in google: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": "Error validating user"})
	}

    // Check if token is expired                                                                                                          
    if time.Now().Unix() > payload.Expires {
        return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Expired credential"})
     }

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string)}

	// Return HTTP 200 if success
	return c.JSON(http.StatusOK, &userProfile)
}

// Creates a ticket from file
func create_ticket(c echo.Context) error {
    file, err := c.FormFile("file")

    if err != nil {
        log.Println("create_ticket - File error", err)
        return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error loading file"})
    }

    // Ensure file has the right format
    format := file.Header["Content-Type"][0]
    if format != "image/png" && format != "image/jpg" && format != "application/pdf" {
        return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Unsupported file format"})
    }

    // TODO: process with Textract
    // TODO: store ticket information

    return c.JSON(http.StatusOK, echo.Map{"message": "Ticket created successfully"})
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

    e := echo.New()
    e.Use(middleware.CORS())
    e.POST("/login/google", login_google)
    e.POST("/ticket", create_ticket)
    e.Logger.Fatal(e.Start(":8000"))
}
