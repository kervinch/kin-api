package xendit

import (
	"fmt"
	"strconv"

	"github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
)

type Xendit struct {
	secretKey string
}

func New(secretKey string) Xendit {
	return Xendit{
		secretKey: secretKey,
	}
}

func (x Xendit) GenerateInvoice(orderID int64, customer xendit.InvoiceCustomer, customAddress xendit.CustomerAddress, invoiceItem []xendit.InvoiceItem, invoiceFee []xendit.InvoiceFee, notificationType []string, total int) (*xendit.Invoice, error) {
	xendit.Opt.SecretKey = x.secretKey

	customerNotificationPreference := xendit.InvoiceCustomerNotificationPreference{
		InvoiceCreated:  notificationType,
		InvoiceReminder: notificationType,
		InvoicePaid:     notificationType,
		InvoiceExpired:  notificationType,
	}

	data := invoice.CreateParams{
		ExternalID:                     strconv.Itoa(int(orderID)),
		Amount:                         float64(total),
		Description:                    "Invoice for product(s) purchase from KIN",
		InvoiceDuration:                86400,
		Customer:                       customer,
		CustomerNotificationPreference: customerNotificationPreference,
		Currency:                       "IDR",
		Items:                          invoiceItem,
		Fees:                           invoiceFee,
	}

	resp, err := invoice.Create(&data)
	if err != nil {
		return nil, err
	}

	fmt.Printf("created invoice: %+v\n", resp)

	return resp, nil
}
