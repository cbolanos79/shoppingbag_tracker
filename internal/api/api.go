package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	model "github.com/cbolanos79/shoppingbag_tracker/internal/model"

	"github.com/golang-jwt/jwt/v5"

	"github.com/labstack/echo/v4"
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

// Check if given google id user exists in database
func FindUserByGoogleID(db *sql.DB, google_uid string) (*model.User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE google_uid = ?", google_uid)

	user := model.User{}
	if err := row.Scan(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Receive credential for Google login and validate it agains Google API
// If credential is valid, extract name and profile picture url
// Else, returns an error
func LoginGoogle(c echo.Context) error {
	login := Login{}
	c.Bind(&login)

	// Return HTTP 422 if credential value is not set
	if len(login.Credential) == 0 {
		log.Println("Missing credential value")
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Missing credential value"})
	}

	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")

	// Validate credential with google client
	payload, err := idtoken.Validate(context.Background(), login.Credential, google_client_id)

	// Return HTTP 422 if there was any error
	if err != nil {
		log.Printf("LoginGoogle - Error validating user in google: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error validating user"})
	}

	// Check if user exists
	db, err := model.NewDB()
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error accessing database"})
	}

	user, err := model.FindUserByGoogleUid(db, payload.Subject)
	if err != nil {
		log.Printf("GoogleLogin - User %s not found, error %v\n", payload.Subject, err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "User not found"})
	}

	// Check if token is expired
	if time.Now().Unix() > payload.Expires {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Expired credential"})
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		Subject:   fmt.Sprint(user.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt_signature := os.Getenv("JWT_SIGNATURE")

	ss, err := token.SignedString([]byte(jwt_signature))
	if err != nil {
		log.Fatal(fmt.Sprintf("Error signing JWT token for user %s", payload.Subject), err.Error())
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error creating token for user"})
	}

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string), ss}

	// Return HTTP 200 if success
	return c.JSON(http.StatusOK, &userProfile)
}

// Check if user from jwt exists or stop if not
func UserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)

		user_id, err := token.Claims.GetSubject()
		if err != nil {
			log.Println("CreateReceipt - Error decoding token\n", err)
			return echo.ErrUnauthorized
		}

		db, err := model.NewDB()
		if err != nil {
			log.Println("CreateReceipt - Error decoding token\n", err)
			return echo.ErrUnauthorized
		}

		user_idd, err := strconv.Atoi(user_id)
		if err != nil {
			log.Println("CreateReceipt - Error decoding token\n", err)
			return echo.ErrUnauthorized
		}

		user, err := model.FindUserById(db, user_idd)
		if user == nil || err != nil {
			log.Println("CreateReceipt - User not found\n", err)
			return echo.ErrUnauthorized
		}

		c.Set("user_id", user)
		return next(c)
	}
}
