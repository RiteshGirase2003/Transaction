
package entity

// Transaction represents a financial transaction between a sender and a receiver.
// It contains all the necessary details related to the transaction including the participants,
// amount, payment methods, status, and timestamp.
//
// Fields:
//   - ID: Unique identifier for the transaction.
//   - SenderID: Identifier for the sender of the transaction.
//   - ReceiverID: Identifier for the receiver of the transaction.
//   - Amount: The amount of money being transferred in the transaction.
//   - PaymentMethod: The payment method used by the sender (e.g., 'UPI', 'CreditCard', 'Bank').
//   - RecievingMethod: The payment method used by the receiver (e.g., 'UPI', 'CreditCard', 'Bank').
//   - SenderPaymentDetails: The payment details of the sender, including UPI, credit card, or bank details.
//   - RecieverPaymentDetails: The payment details of the receiver, including UPI, credit card, or bank details.
//   - Status: The current status of the transaction (e.g., 'success', 'failed', 'pending', 'cancelled').
//   - Timestamp: The time when the transaction occurred.
//   - TransactionType: The type of the transaction (e.g., 'transfer', 'payment').
//   - ActionBy: Identifier of the person performing the action on the transaction (optional).
type Transaction struct {
	ID                     string         `json:"id"`
	SenderID               string         `json:"sender_id" validate:"required"`

	// This is Receiver ID  ----
	ReceiverID             string         `json:"receiver_id,omitempty"`
	Amount                 float64        `json:"amount" validate:"required,gt=0"`
	PaymentMethod          string         `json:"sender_payment_method"`
	RecievingMethod        string         `json:"reciever_payment_method"`
	SenderPaymentDetails   PaymentDetails `json:"sernder_payment_details"`
	RecieverPaymentDetails PaymentDetails `json:"reciever_payment_details"`
	Status                 string         `json:"status" validate:"required"`
	Timestamp              int64          `json:"timestamp"`
	TransactionType        string         `json:"transaction_type" validate:"required"`
	ActionBy               string         `json:"action_by,omitempty"`
}

// PaymentDetails contains the payment information for both sender and receiver.
// It includes details about the sender's and receiver's UPI, credit card, or bank account information.
type PaymentDetails struct {
	UPI         UPIDetails        `json:"upi,omitempty"`
	CreditCard  CreditCardDetails `json:"credit_card,omitempty"`
	BankDetails BankDetails       `json:"bank_details,omitempty"`
}

// UPIDetails represents the UPI details for the sender or receiver.
// It contains the UPI ID.
type UPIDetails struct {
	UpiId string `json:"upi_id"`
}

// CreditCardDetails represents the credit card details for the sender or receiver.
// It contains the card ID and the last four digits of the card number.
type CreditCardDetails struct {
	CardID         string `json:"card_id"`
	LastFourNumber string `json:"card_number"`
}

// BankDetails represents the bank account details for the sender or receiver.
// It includes the account number, IFSC code, and name of the bank.
type BankDetails struct {
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"`
	BankName      string `json:"bank_name"`
}

// RequestBody represents the structure of the request body for initiating a transaction.
// It includes sender and receiver details, payment methods, and payment details for both participants.
type RequestBody struct {
	SenderID               string         `json:"sender_id"`
	ReceiverID             string         `json:"receiver_id,omitempty"`
	Amount                 float64        `json:"amount"`
	PaymentMethod          string         `json:"payment_method"`
	RecievingMethod        string         `json:"recieving_method"`
	TransactionType        string         `json:"transaction_type"`
	SenderPaymentDetails   PaymentDetails `json:"sender_payment_details"`
	ReceiverPaymentDetails PaymentDetails `json:"receiver_payment_details"`
}

// MakePaymentRequest represents the structure for a request to make a payment.
// It includes sender and receiver details, amount, payment methods, and payment details.
type MakePaymentRequest struct {
	RequesterID             string         `json:"requester_id" validate:"required"`
	PayerID                 string         `json:"payer_id" validate:"required"`
	Amount                  float64        `json:"amount" validate:"required,gt=0"`
	RequesterPaymentMethod  string         `json:"requester_payment_method" validate:"required,eq=upi"`
	PayerPaymentMethod      string         `json:"payer_payment_method" validate:"required,eq=upi"`
	TransactionType         string         `json:"transaction_type"`
	RequesterPaymentDetails PaymentDetails `json:"requester_payment_details" validate:"required"`
	PayerPaymentDetails     PaymentDetails `json:"payer_payment_details" validate:"required"`
}

// PaymentRequestAction represents the structure for an action to be performed on a payment request.
// It includes the request ID and the action to be performed (e.g., "approve", "reject").
type PaymentRequestAction struct {
	RequestID string `json:"request_id" validate:"required"`
	Action    string `json:"action" validate:"required"`
}

// TransactionRequest represents the structure for a transaction request.
// It includes the transaction ID, sender and receiver account numbers, amount, payment method, and other relevant details.
type TransactionRequest struct {
	ID             string  `json:"id"`
	RequesterAccNo string  `json:"requesterAccNo"`
	PayerAccNo     string  `json:"payerAccNo"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"paymentMethod"`
	TransactionID  string  `json:"transactionID"`
	From           string  `json:"from"`
	To             string  `json:"to"`
}
