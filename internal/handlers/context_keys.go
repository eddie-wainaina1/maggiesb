package handlers

// Context keys for dependency injection in tests
const (
	ContextKeyOrderRepo    = "test_order_repo"
	ContextKeyInvoiceRepo  = "test_invoice_repo"
	ContextKeyUserRepo     = "test_user_repo"
	ContextKeyProductRepo  = "test_product_repo"
	ContextKeyPaymentRepo  = "test_payment_repo"
	ContextKeyReversalRepo = "test_reversal_repo"
	ContextKeyMpesaClient  = "test_mpesa_client"
)
