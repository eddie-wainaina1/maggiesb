package models

// MpesaPaymentRequest payload to initiate M-Pesa payment
type MpesaPaymentRequest struct {
	InvoiceID string `json:"invoiceId" binding:"required"`
	Phone     string `json:"phone" binding:"required,min=10"`
}

// MpesaPaymentResponse from STK Push initiation
type MpesaPaymentResponse struct {
	CheckoutRequestID string `json:"checkoutRequestId"`
	ResponseCode      string `json:"responseCode"`
	ResponseMessage   string `json:"responseMessage"`
	CustomerMessage   string `json:"customerMessage"`
}

// MpesaCallback represents incoming M-Pesa payment callback
type MpesaCallback struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// PaymentRecord stores M-Pesa payment transaction details
type PaymentRecord struct {
	ID                 string `bson:"_id" json:"id"`
	InvoiceID          string `bson:"invoiceId" json:"invoiceId"`
	OrderID            string `bson:"orderId" json:"orderId"`
	CheckoutRequestID  string `bson:"checkoutRequestId" json:"checkoutRequestId"`
	MerchantRequestID  string `bson:"merchantRequestId" json:"merchantRequestId"`
	Phone              string `bson:"phone" json:"phone"`
	Amount             float64 `bson:"amount" json:"amount"`
	MpesaReceiptNumber string `bson:"mpesaReceiptNumber" json:"mpesaReceiptNumber"`
	TransactionDate    string `bson:"transactionDate" json:"transactionDate"`
	Status             string `bson:"status" json:"status"` // "initiated", "completed", "failed"
	CreatedAt          string `bson:"createdAt" json:"createdAt"`
	UpdatedAt          string `bson:"updatedAt" json:"updatedAt"`
}
