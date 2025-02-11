package utils

import (
	"encoding/json"
	"errors"
	"go-transaction/entity"
	"net/http"
	"strings"
)

// ReadRequestBody decodes the request body into a RequestBody object
// and validates the required fields and payment details.
func ReadRequestBody(req *http.Request, data *entity.RequestBody) error {
	// Decode JSON body
	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		return err
	}

	if strings.EqualFold(data.TransactionType, "Payment") && data.TransactionType != ""{
		return errors.New("Transaction Type must be Payment")
	}

	// Validate required fields
	if data.SenderID == "" || data.PaymentMethod == "" || data.RecievingMethod == "" {
		return errors.New("SenderID, PaymentMethod, and RecievingMethod are required")
	}

	// Validate Sender Payment Details
	if err := validatePaymentDetails(data.PaymentMethod, data.SenderPaymentDetails); err != nil {
		return err
	}

	// Validate Receiver Payment Details
	if err := validatePaymentDetails(data.RecievingMethod, data.ReceiverPaymentDetails); err != nil {
		return err
	}

	return nil
}

// validatePaymentDetails checks the validity of the payment details based on the payment method.
func validatePaymentDetails(method string, details entity.PaymentDetails) error {
	switch method {
	case "UPI":
		if details.UPI.UpiId == "" {
			return errors.New("UPI ID is required for UPI transactions")
		}
	case "CREDIT_CARD":
		if details.CreditCard.CardID == "" || details.CreditCard.LastFourNumber == "" {
			return errors.New("Both Card ID and Last Four Digits are required for Credit Card transactions")
		}
	case "BANK":
		if details.BankDetails.AccountNumber == "" || details.BankDetails.IFSCCode == "" || details.BankDetails.BankName == "" {
			return errors.New("Account Number, IFSC Code, and Bank Name are required for Bank transactions")
		}
	default:
		return errors.New("Invalid Payment or Receiving Method")
	}
	return nil
}

// ReadMakePaymentRequest decodes the request body into a MakePaymentRequest object
// and validates the required fields and UPI details for a payment transaction.
func ReadMakePaymentRequest(req *http.Request, data *entity.MakePaymentRequest) error {
	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		return err
	}


	if strings.EqualFold(data.TransactionType, "Request") && data.TransactionType != "" {
		return errors.New("Transaction Type must be Request")
	}

	if data.RequesterID == "" || data.PayerID == "" || data.PayerPaymentMethod == "" || data.RequesterPaymentMethod == "" {
		return errors.New("SenderID, ReceiverID, PaymentMethod, and RecievingMethod are required")
	}

	if data.Amount <= 0 {
		return errors.New("Amount must be greater than 0")
	}

	if !strings.EqualFold(data.RequesterPaymentMethod, "upi") || !strings.EqualFold(data.PayerPaymentMethod, "upi") {
		return errors.New("Only UPI is supported for PaymentMethod and RecievingMethod")
	}
	

	if err := validateUPIDetails(data.RequesterPaymentDetails.UPI); err != nil {
		return errors.New("Invalid sender UPI details: " + err.Error())
	}

	if err := validateUPIDetails(data.PayerPaymentDetails.UPI); err != nil {
		return errors.New("Invalid receiver UPI details: " + err.Error())
	}

	return nil
}

// ReadPaymentRequestAction decodes the request body into a PaymentRequestAction object
// and validates the required fields and action type for a transaction request.
func ReadPaymentRequestAction(req *http.Request, data *entity.PaymentRequestAction) error {
	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		return err
	}

	if data.RequestID == "" || data.Action == "" {
		return errors.New("TransactionID and UserID are required")
	}

	if !strings.EqualFold(data.Action, "Accept") && !strings.EqualFold(data.Action, "Cancel") {
		return errors.New("Invalid action. Action must be either 'Accept' or 'Reject'")
	}

	return nil
}

// validateUPIDetails checks the validity of UPI ID details.
func validateUPIDetails(upi entity.UPIDetails) error {
	if strings.TrimSpace(upi.UpiId) == "" {
		return errors.New("UPI ID is required")
	}
	return nil
}
