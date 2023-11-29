package main

import (
	"log"
	"os"

	"github.com/cbolanos79/shoppingbag_tracker/internal/api"
	"github.com/cbolanos79/shoppingbag_tracker/internal/model"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")
	if len(google_client_id) == 0 {
		log.Fatal("Empty value for GOOGLE_CLIENT_ID")
	}

	db_name := os.Getenv("DB_NAME")
	if len(db_name) == 0 {
		log.Fatal("Empty value for DB_NAME")
	}

	_, err = model.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	jwt_signature := os.Getenv("JWT_SIGNATURE")
	if len(jwt_signature) == 0 {
		log.Fatal("Missing jwt signature")
	}

	e := echo.New()
	e.Use(middleware.CORS())
	e.POST("/login/google", api.LoginGoogle)
	e.Logger.Fatal(e.Start(":8000"))
}
