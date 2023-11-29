package receipt_scanner

import (
	mime "mime/multipart"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/cbolanos79/shoppingbag_tracker/internal/model"
)

func NewAwsSession() (*session.Session, error) {
	aws_session, err := session.NewSessionWithOptions(session.Options{
		Profile: "textract",
		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-west-1"),
		},
	})

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

	// Get total amount from receipt
	total, err := strconv.ParseFloat(strings.Replace(SearchExpense(res.ExpenseDocuments[0].SummaryFields, "TOTAL"), ",", ".", -1), 64)
	if err != nil {
		total = -1
	}

	receipt.Total = total

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
		receipt.Items = append(receipt.Items, model.ReceiptItem{Name: name, Quantity: quantity, Price: price, UnitPrice: unit_price})
	}

	return receipt, nil
}
