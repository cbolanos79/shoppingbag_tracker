package main

import (
	"fmt"
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

	s, err := model.NewDBStorage()
	if err != nil {
		log.Fatal(err)
	}

	jwt_signature := os.Getenv("JWT_SIGNATURE")
	if len(jwt_signature) == 0 {
		log.Fatal("Missing jwt signature")
	}

	s.InitDB()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	api.RegisterRoutes(s, e, jwt_signature)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
