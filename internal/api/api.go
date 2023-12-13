package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"strconv"
	"time"

	model "github.com/cbolanos79/shoppingbag_tracker/internal/model"
	"github.com/cbolanos79/shoppingbag_tracker/internal/receipt_scanner"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"

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

type ErrorMessage struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type Handler struct {
	s model.IStorage
}

func RegisterRoutes(s model.IStorage, e *echo.Echo, jwt_signature string) {

	a := Api{
		s: s,
	}

	e.POST("/analyze", a.AnalyzeReceipt, echojwt.JWT([]byte(jwt_signature)), a.UserMiddleware)
	e.POST("/receipt", a.CreateReceipt, echojwt.JWT([]byte(jwt_signature)), a.UserMiddleware)
	e.POST("/login/google", a.LoginGoogle)
}

// Receive credential for Google login and validate it agains Google API
// If credential is valid, extract name and profile picture url
// Else, returns an error
func (h *Handler) LoginGoogle(c echo.Context) error {
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
		log.Printf("LoginGoogle - Error validating user in google: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error validating user", []string{err.Error()}})
	}

	// Check if user exists
	user, err := h.s.FindUserByGoogleUid(payload.Subject)
	if err != nil {
		log.Printf("GoogleLogin - User %s not found, error %v", payload.Subject, err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"User not found", []string{err.Error()}})
	}

	// Check if token is expired
	if time.Now().Unix() > payload.Expires {
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Expired credential", []string{err.Error()}})
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		Subject:   fmt.Sprint(user.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt_signature := os.Getenv("JWT_SIGNATURE")

	ss, err := token.SignedString([]byte(jwt_signature))
	if err != nil {
		log.Printf("Error signing JWT token for user %s: %v", payload.Subject, err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error creating token for user", []string{err.Error()}})
	}

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string), ss}

	// Return HTTP 200 if success
	return c.JSON(http.StatusOK, &userProfile)
}

// Analyze a receipt image and returns information in json format or error if could not be analyzed
func (h *Handler) AnalyzeReceipt(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("CreateReceipt - Error processing form file: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error opening file", []string{err.Error()}})
	}

	scanner, err := receipt_scanner.NewScanner()
	if err != nil {
		log.Printf("CreateReceipt - Error initializing scanner: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error initializing scanner", []string{err.Error()}})
	}

	f, err := file.Open()
	if err != nil {
		log.Printf("CreateReceipt - Error opening file: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error opening file", []string{err.Error()}})
	}

	receipt, err := scanner.Scan(f, file.Size)
	if err != nil {
		log.Printf("CreateReceipt - Error opening file: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error analyzing file", []string{err.Error()}})
	}

	return c.JSON(http.StatusOK, echo.Map{"receipt": receipt})
}

// Create a receipt from given file using a valid user, or return error with status 422 if can not create
func (h *Handler) CreateReceipt(c echo.Context) error {

	user := c.Get("user_id").(*model.User)
	var receipt model.Receipt

	err := c.Bind(&receipt)
	if err != nil {
		log.Printf("CreateReceipt - Error parsing JSON: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error parsing JSON", []string{err.Error()}})
	}

	receipt.UserID = user.ID

	_, err = h.s.CreateReceipt(&receipt)
	if err != nil {
		log.Printf("CreateReceipt - Error creating receipt: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error creating receipt", []string{err.Error()}})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Receipt created successfully", "receipt": receipt})
}

// Check if user from jwt exists or stop if not
func (h *Handler) UserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)

		user_id, err := token.Claims.GetSubject()
		if err != nil {
			log.Printf("CreateReceipt - Error decoding token: %v", err)
			return echo.ErrUnauthorized
		}

		user_idd, err := strconv.Atoi(user_id)
		if err != nil {
			log.Printf("CreateReceipt - Error decoding token: %v", err)
			return echo.ErrUnauthorized
		}

		user, err := h.s.FindUserById(user_idd)
		if user == nil || err != nil {
			log.Printf("CreateReceipt - User not found: %v", err)
			return echo.ErrUnauthorized
		}

		c.Set("user_id", user)
		return next(c)
	}
}
