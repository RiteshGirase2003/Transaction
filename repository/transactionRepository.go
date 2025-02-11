package repository

import (
	"context"
	"fmt"
	"go-transaction/entity"

	"github.com/rs/zerolog/log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// GetAccNo retrieves the account number associated with the given payment method and details.
//
// Parameters:
//   - ctx: The context for Firestore operations.
//   - client: Firestore client to interact with the database.
//   - payment_method: The type of payment method (e.g., "upi_id", "card_id", "account_number").
//   - details: The identifier for the payment method (e.g., actual UPI ID, card ID, or account number).
//
// Returns:
//   - The corresponding account number if found.
//   - An error if no matching document is found or if an issue occurs during retrieval.
func GetAccNo(ctx context.Context, client *firestore.Client, payment_method, details string) (string, error) {
	bankDetailsRef := client.Collection("BankDetails")

	// Query Firestore for a document matching the given payment method and details
	senderQuery := bankDetailsRef.Where(payment_method, "==", details).Documents(ctx)
	senderDoc, err := senderQuery.Next()
	if err != nil {
		if err == iterator.Done {
			log.Error().
				Str("sender "+payment_method, details).
				Msg("No matching document found for sender")
			return "", fmt.Errorf("no matching document found for sender %s: %s", payment_method, details)
		}
		log.Error().
			Err(err).
			Msg("Error fetching sender document")
		return "", fmt.Errorf("failed to fetch sender document: %v", err)
	}

	// Extract the account number from the sender document
	userData := senderDoc.Data()
	AccNo, ok := userData["account_number"].(string)
	if !ok {
		log.Error().Msg("Sender account number is not of type string")
		return "", fmt.Errorf("invalid sender account number format")
	}

	return AccNo, nil
}

// GetUserAccNo retrieves the account numbers for both sender and receiver based on their payment methods.
//
// Parameters:
//   - ctx: The context for Firestore operations.
//   - client: Firestore client to interact with the database.
//   - paymentMethod: The sender's payment method (e.g., "UPI", "BANK", "CREDIT_CARD").
//   - receivingMethod: The receiver's payment method (e.g., "UPI", "BANK", "CREDIT_CARD").
//   - paymentDetails: The sender's payment details containing relevant payment method identifiers.
//   - receivingDetails: The receiver's payment details containing relevant payment method identifiers.
//
// Returns:
//   - Sender's account number.
//   - Receiver's account number.
//   - An error if account retrieval fails.
func GetUserAccNo(ctx context.Context, client *firestore.Client, paymentMethod, receivingMethod string, paymentDetails entity.PaymentDetails, receivingDetails entity.PaymentDetails) (string, string, error) {
	switch {
	case paymentMethod == "UPI" && receivingMethod == "UPI":
		senderAccNo, err := GetAccNo(ctx, client, "upi_id", paymentDetails.UPI.UpiId)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch sender account details")
		}
		receiverAccNo, err := GetAccNo(ctx, client, "upi_id", receivingDetails.UPI.UpiId)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch receiver account details")
		}
		return senderAccNo, receiverAccNo, nil

	case paymentMethod == "UPI" && receivingMethod == "BANK":
		senderAccNo, err := GetAccNo(ctx, client, "upi_id", paymentDetails.UPI.UpiId)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch sender account details")
		}
		receiverAccNo, err := GetAccNo(ctx, client, "account_number", receivingDetails.BankDetails.AccountNumber)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch receiver account details")
		}
		return senderAccNo, receiverAccNo, nil

	case paymentMethod == "CREDIT_CARD" && receivingMethod == "UPI":
		senderAccNo, err := GetAccNo(ctx, client, "card_id", paymentDetails.CreditCard.CardID)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch sender account details")
		}
		receiverAccNo, err := GetAccNo(ctx, client, "upi_id", receivingDetails.UPI.UpiId)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch receiver account details")
		}
		return senderAccNo, receiverAccNo, nil

	case paymentMethod == "BANK" && receivingMethod == "BANK":
		senderAccNo, err := GetAccNo(ctx, client, "account_number", paymentDetails.BankDetails.AccountNumber)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch sender account details")
		}
		receiverAccNo, err := GetAccNo(ctx, client, "account_number", receivingDetails.BankDetails.AccountNumber)
		if err != nil {
			return "", "", fmt.Errorf("unable to fetch receiver account details")
		}
		return senderAccNo, receiverAccNo, nil

	default:
		log.Error().
			Str("paymentMethod", paymentMethod).
			Str("receivingMethod", receivingMethod).
			Msg("Invalid payment method combination")
		return "", "", fmt.Errorf("invalid payment method or receiving method")
	}
}
