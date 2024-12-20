package services

import (
	"fmt"
	xendit "github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
	"os"
)

type PaymentService struct{}

// Invoice adalah struct untuk response payment
type Invoice struct {
	ID          string  `json:"id"`
	ExternalID  string  `json:"external_id"`
	Amount      float64 `json:"amount"`
	PayerEmail  string  `json:"payer_email"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	InvoiceURL  string  `json:"invoice_url"`
}

func NewPaymentService() *PaymentService {
	xendit.Opt.SecretKey = os.Getenv("XENDIT_SECRET_KEY")
	return &PaymentService{}
}

func (s *PaymentService) CreatePayment(userEmail string, amount float64, rentalID uint) (*Invoice, error) {
	// Buat parameter untuk invoice xendit
	params := invoice.CreateParams{
		ExternalID:  fmt.Sprintf("rental-%d", rentalID),
		Amount:      amount,
		PayerEmail:  userEmail,
		Description: "Car Rental Payment",
	}

	// Buat invoice di xendit
	resp, err := invoice.Create(&params)
	if err != nil {
		return nil, err
	}

	// Convert ke struct Invoice kita
	invoice := &Invoice{
		ID:          resp.ID,
		ExternalID:  resp.ExternalID,
		Amount:      resp.Amount,
		PayerEmail:  resp.PayerEmail,
		Description: resp.Description,
		Status:      resp.Status,
		InvoiceURL:  resp.InvoiceURL,
	}

	return invoice, nil
}
