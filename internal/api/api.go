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
	"github.com/relvacode/iso8601"

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

type ErrorMessage struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
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
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error validating user", []string{err.Error()}})
	}

	// Check if user exists
	db, err := model.NewDB()
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error accessing database", []string{err.Error()}})
	}
	defer db.Close()

	user, err := model.FindUserByGoogleUid(db, payload.Subject)
	if err != nil {
		log.Printf("GoogleLogin - User %s not found, error %v\n", payload.Subject, err)
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
		log.Fatal(fmt.Sprintf("Error signing JWT token for user %s", payload.Subject), err.Error())
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error creating token for user", []string{err.Error()}})
	}

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string), ss}

	// Return HTTP 200 if success
	return c.JSON(http.StatusOK, &userProfile)
}

// Create a receipt from given file using a valid user, or return error with status 422 if can not create
// Receipt is analyzed by Textract, and then store results into database
func CreateReceipt(c echo.Context) error {
	// By default, token is stored in user key

	file, err := c.FormFile("file")
	if err != nil {
		log.Println("CreateReceipt - Error processing form file\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error opening file", []string{err.Error()}})
	}

	session, err := receipt_scanner.NewAwsSession()
	if err != nil {
		log.Println("CreateReceipt - Error creating new aws session\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error connecting to aws", []string{err.Error()}})
	}

	db, err := model.NewDB()
	if err != nil {
		log.Println("CreateReceipt - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error connecting to database", []string{err.Error()}})
	}
	defer db.Close()

	f, err := file.Open()
	if err != nil {
		log.Println("CreateReceipt - Error opening file\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error opening file", []string{err.Error()}})
	}

	receipt, err := receipt_scanner.Scan(session, f, file.Size)
	if err != nil {
		log.Println("CreateReceipt - Error opening file\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error analyzing file", []string{err.Error()}})
	}

	user := c.Get("user_id").(*model.User)
	receipt.UserID = user.ID

	_, err = model.CreateReceipt(db, receipt)
	if err != nil {
		log.Println("CreateReceipt - Error creating receipt\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error creating receipt", []string{err.Error()}})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Receipt created successfully", "receipt": receipt})
}

// Return list of receipts for current user
func GetReceipts(c echo.Context) error {

	db, err := model.NewDB()
	if err != nil {
		log.Println("GetReceipts - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error connecting to database", []string{err.Error()}})
	}
	defer db.Close()

	user := c.Get("user_id").(*model.User)

	var filters model.ReceiptFilter

	// Supermarket filter
	filters.Supermarket = c.QueryParam("supermarket")

	page := c.QueryParam("page")
	per_page := c.QueryParam("per_page")
	min_date := c.QueryParam("min_date")
	max_date := c.QueryParam("max_date")
	filters.Item = c.QueryParam("item")

	// Page filter
	if len(page) > 0 && len(per_page) > 0 {
		filters.Page, err = strconv.ParseInt(page, 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error in page param format", []string{err.Error()}})
		}

		filters.PerPage, err = strconv.ParseInt(per_page, 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error in per_page param format", []string{err.Error()}})
		}
	}

	// Minimum date
	if len(min_date) > 0 {
		// Parse ISO8601 format
		tmin_date, err := iso8601.ParseString(min_date)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error in min_date param format", []string{err.Error()}})
		}

		filters.MinDate = &tmin_date

		// Maximum date
		if len(max_date) > 0 && filters.MinDate != nil {
			// Parse ISO8601 format
			tmax_date, err := iso8601.ParseString(max_date)
			if err != nil {
				return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error in max_date param format", []string{err.Error()}})
			}

			filters.MaxDate = &tmax_date
		}
	}

	receipts, err := model.FindAllReceiptsForUser(db, user, &filters)
	if err != nil {
		log.Println("GetReceipts - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error getting receipts list", []string{err.Error()}})
	}

	return c.JSON(http.StatusOK, echo.Map{"receipts": receipts})
}

// Return list of items for given receipt owned by user
func GetReceipt(c echo.Context) error {

	db, err := model.NewDB()
	if err != nil {
		log.Println("GetReceipt - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error connecting to database", []string{err.Error()}})
	}
	defer db.Close()

	user := c.Get("user_id").(*model.User)
	receipt_id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		log.Println("GetReceipt - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error connecting to database", []string{err.Error()}})
	}

	receipt, err := model.FindReceiptForUser(db, int(receipt_id), int(user.ID))
	if err != nil {
		log.Println("GetReceipt - Error connecting to database\n", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorMessage{"Error getting receipts list", []string{err.Error()}})
	}

	return c.JSON(http.StatusOK, echo.Map{"receipt": receipt})
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
		defer db.Close()

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
