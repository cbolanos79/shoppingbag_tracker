package receipt_scanner

import (
	mime "mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/cbolanos79/shoppingbag_tracker/internal/model"
)

func NewAwsSession() (*session.Session, error) {
	aws_session, err := session.NewSession()

	if err != nil {
		return nil, err
	}

	return aws_session, nil
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

// Auxiliar function to search string into an array of textract.ExpenseField
func SearchCurrency(item []*textract.ExpenseField) string {
	for _, item := range item {
		if *item.Type.Text == "TOTAL" {
			return *item.Currency.Code
		}
	}
	return ""
}

// Analyze ticket on Textract using OCR and AI, and get in response structured information about receipt
func Scan(aws_session *session.Session, file mime.File, size int64) (*model.Receipt, error) {

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
	receipt := &model.Receipt{}
	receipt.Supermarket = sres[0]

	receipt_date := strings.Replace(SearchExpense(res.ExpenseDocuments[0].SummaryFields, "INVOICE_RECEIPT_DATE"), ",", ".", -1)
	var date time.Time

	// Sometimes, a receipt can have date with format dd.mm.yy due bad quality image or any other problems, which can be a problem to parse
	// Therefore, check if date has this format and parse with the right layout
	pattern := regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{1,2}$`)

	if pattern.FindIndex([]byte(receipt_date)) == nil {
		receipt_date = strings.Replace(receipt_date, "-", "/", -1)

		date, err = time.Parse("02/01/2006", receipt_date)

		if err != nil {
			return nil, err
		}
	} else {
		date, err = time.Parse("02.01.06", receipt_date)

		if err != nil {
			return nil, err
		}
	}

	receipt.Date = date

	// Get total amount from receipt
	total, err := strconv.ParseFloat(strings.Replace(SearchExpense(res.ExpenseDocuments[0].SummaryFields, "TOTAL"), ",", ".", -1), 64)
	if err != nil {
		total = -1
	}
	receipt.Total = total

	// Get currency
	receipt.Currency = SearchCurrency(res.ExpenseDocuments[0].SummaryFields)

	// Iterate over each concept from receipt
	for _, line_item := range res.ExpenseDocuments[0].LineItemGroups[0].LineItems {
		name := SearchExpense(line_item.LineItemExpenseFields, "ITEM")

		squantity := SearchExpense(line_item.LineItemExpenseFields, "QUANTITY")
		quantity := 1.0

		// Some receipts have not quantity field, therefore set 1 by default
		if len(squantity) > 0 {
			quantity, err = strconv.ParseFloat(strings.Replace(squantity, ",", ".", -1), 64)
			if err != nil {
				quantity = -1
			}
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
		receipt.Items = append(receipt.Items, model.ReceiptItem{Name: name, Quantity: quantity, Price: price, UnitPrice: unit_price})
	}

	return receipt, nil
}
