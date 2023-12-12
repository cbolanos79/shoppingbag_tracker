package receipt_scanner

import (
	"fmt"
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
			currency := item.Currency
			if currency == nil {
				return ""
			} else {
				return *currency.Code
			}
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
	stotal := SearchExpense(res.ExpenseDocuments[0].SummaryFields, "TOTAL")
	amount_exp := regexp.MustCompile(`\d+(\,|\.)\d+`)

	total := amount_exp.Find([]byte(stotal))
	if total == nil {
		return nil, fmt.Errorf("error parsing total amount: %s", stotal)
	}

	receipt.Total, err = strconv.ParseFloat(strings.Replace(string(total), ",", ".", -1), 64)

	if err != nil {
		return nil, err
	}

	// Get currency
	receipt.Currency = SearchCurrency(res.ExpenseDocuments[0].SummaryFields)

	// Iterate over each concept from receipt
	for index, line_item := range res.ExpenseDocuments[0].LineItemGroups[0].LineItems {
		name := SearchExpense(line_item.LineItemExpenseFields, "ITEM")

		squantity := SearchExpense(line_item.LineItemExpenseFields, "QUANTITY")
		quantity := 1.0

		// Some receipts have not quantity field, therefore set 1 by default
		if len(squantity) > 0 {
			quantity, err = strconv.ParseFloat(strings.Replace(squantity, ",", ".", -1), 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing quantity: %v", err)
			}
		}

		sprice := SearchExpense(line_item.LineItemExpenseFields, "PRICE")
		var price float64
		if len(sprice) > 0 {
			price, err = strconv.ParseFloat(strings.Replace(sprice, ",", ".", -1), 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing price: %v", err)
			}
		} else {
			return nil, fmt.Errorf("empty price for item #%d", index)
		}

		sunit_price := SearchExpense(line_item.LineItemExpenseFields, "UNIT_PRICE")
		var unit_price float64
		if len(sunit_price) > 0 {
			// Sometimes, unit price can have alphanumeric information
			sunit_price := amount_exp.Find([]byte(sunit_price))
			if sunit_price == nil {
				return nil, fmt.Errorf("error parsing unit price: invalid format %s", sunit_price)
			}
			unit_price, err = strconv.ParseFloat(strings.Replace(string(sunit_price), ",", ".", -1), 64)

			if err != nil {
				return nil, fmt.Errorf("error parsing unit price: %v", err)
			}
		}

		// Add each item to receipt
		receipt.Items = append(receipt.Items, model.ReceiptItem{Name: name, Quantity: quantity, Price: price, UnitPrice: unit_price})
	}

	return receipt, nil
}
