package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	mime "mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
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

type TicketItem struct {
	ID        int64
	Name      string
	Quantity  int64
	Price     float64
	UnitPrice float64
}

type Ticket struct {
	ID          int64
	Supermarket string
	Total       float64
	Items       []TicketItem
}

var google_client_id string
var db *sql.DB
var jwt_signature string
var aws_session *session.Session

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

// Auxiliar function to search string into an array of textract.ExpenseField
func SearchExpense(item []*textract.ExpenseField, s string) string {
	for _, item := range item {
		if *item.Type.Text == s {
			return *item.ValueDetection.Text
		}
	}
	return ""
}

// Analyze ticket on Textract using OCR and AI, and get in response structured information about receipt
func ParseReceipt(file mime.File, size int64) (*Ticket, error) {

	// Create object to e
	svc := textract.New(aws_session)

	// Allocate enough space to read file
	b := make([]byte, size)
	_, err := file.Read(b)

	if err != nil {
		return nil, err
	}

	// Make request to Textract in order to analyze data
	res, err := svc.AnalyzeExpense(&textract.AnalyzeExpenseInput{
		Document: &textract.Document{
			Bytes: b,
		},
	})

	if err != nil {
		return nil, err
	}

	// Get supermarket name
	s := *res.ExpenseDocuments[0].SummaryFields[0].ValueDetection.Text
	sres := strings.Split(s, "\n")
	ticket := &Ticket{}
	ticket.Supermarket = sres[0]

	// Get total amount from receipt
	total, err := strconv.ParseFloat(strings.Replace(SearchExpense(res.ExpenseDocuments[0].SummaryFields, "TOTAL"), ",", ".", -1), 64)
	if err != nil {
		total = -1
	}

	ticket.Total = total

	// Iterate over each concept from receipt
	for _, line_item := range res.ExpenseDocuments[0].LineItemGroups[0].LineItems {
		name := SearchExpense(line_item.LineItemExpenseFields, "ITEM")

		quantity, err := strconv.ParseInt(SearchExpense(line_item.LineItemExpenseFields, "QUANTITY"), 10, 64)
		if err != nil {
			quantity = -1
		}

		price, err := strconv.ParseFloat(strings.Replace(SearchExpense(line_item.LineItemExpenseFields, "PRICE"), ",", ".", -1), 64)
		if err != nil {
			price = -1
		}

		unit_price, err := strconv.ParseFloat(strings.Replace(SearchExpense(line_item.LineItemExpenseFields, "UNIT_PRICE"), ",", ".", -1), 64)
		if err != nil {
			unit_price = -1
		}

		// Add each item to receipt
		ticket.Items = append(ticket.Items, TicketItem{Name: name, Quantity: quantity, Price: price, UnitPrice: unit_price})
	}

	return ticket, nil
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
func LoginGoogle(c echo.Context) error {
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

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		Subject:   payload.Subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(jwt_signature))
	if err != nil {
		log.Fatal(fmt.Sprintf("Error signing JWT token for user %s", payload.Subject), err.Error())
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"message": "Error creating token for user"})
	}

	userProfile := UserProfile{payload.Claims["name"].(string), payload.Claims["picture"].(string), ss}

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

	aws_session, err = session.NewSessionWithOptions(session.Options{
		Profile: "textract",
		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-west-1"),
		},
	})

	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.CORS())
	e.POST("/login/google", LoginGoogle)
	e.Logger.Fatal(e.Start(":8000"))
}
