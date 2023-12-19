package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cbolanos79/shoppingbag_tracker/internal/api"
	"github.com/cbolanos79/shoppingbag_tracker/internal/model"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
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

	db, err := model.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	jwt_signature := os.Getenv("JWT_SIGNATURE")
	if len(jwt_signature) == 0 {
		log.Fatal("Missing jwt signature")
	}

	model.InitDB(db)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.GET("/receipts/:id", api.GetReceipt, echojwt.JWT([]byte(jwt_signature)), api.UserMiddleware)
	e.GET("/receipts", api.GetReceipts, echojwt.JWT([]byte(jwt_signature)), api.UserMiddleware)

	e.POST("/receipt", api.CreateReceipt, echojwt.JWT([]byte(jwt_signature)), api.UserMiddleware)
	e.POST("/login/google", api.LoginGoogle)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
