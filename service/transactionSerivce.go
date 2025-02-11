package service

import (
	"context"
	"fmt"
	"go-transaction/config"
	"go-transaction/entity"
	"go-transaction/repository"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

var transactionPool = sync.Pool{
	New: func() interface{} {
		return &entity.Transaction{}
	},
}

var transactionRequestPool = sync.Pool{
	New: func() interface{} {
		return &entity.TransactionRequest{}
	},
}

func resetTransaction(t *entity.Transaction) {
	t.ID = ""
	t.SenderID = ""
	t.ReceiverID = ""
	t.ActionBy = ""
	t.Amount = 0
	t.PaymentMethod = ""
	t.RecievingMethod = ""
	t.SenderPaymentDetails = entity.PaymentDetails{}
	t.RecieverPaymentDetails = entity.PaymentDetails{}
	t.Status = ""
	t.Timestamp = 0
	t.TransactionType = ""
}

func resetTransactionRequest(t *entity.TransactionRequest) {
	t.ID = ""
	t.RequesterAccNo = ""
	t.PayerAccNo = ""
	t.Amount = 0
	t.PaymentMethod = ""
	t.TransactionID = ""
	t.From = ""
	t.To = ""
}

func InitiateTransaction(ctx context.Context, requestBody entity.RequestBody) error {
	transaction := transactionPool.Get().(*entity.Transaction)
	defer func() {
		resetTransaction(transaction)
		transactionPool.Put(transaction)
	}()

	transaction.SenderID = requestBody.SenderID
	transaction.Amount = requestBody.Amount
	transaction.PaymentMethod = strings.ToUpper(requestBody.PaymentMethod)
	transaction.RecievingMethod = strings.ToUpper(requestBody.RecievingMethod)
	transaction.SenderPaymentDetails = requestBody.SenderPaymentDetails
	transaction.RecieverPaymentDetails = requestBody.ReceiverPaymentDetails
	transaction.Status = "pending"
	transaction.TransactionType = requestBody.TransactionType
	transaction.ActionBy = requestBody.SenderID
	transaction.Timestamp = time.Now().Unix()

	app, err := config.InitFirebase()
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to initialize Firebase app")
		return err
	}

	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get Firestore client")
		return err
	}
	defer client.Close()

	senderAccNo, receiverAccNo, err := repository.GetUserAccNo(ctx, client,
		strings.ToUpper(requestBody.PaymentMethod),
		strings.ToUpper(requestBody.RecievingMethod),
		requestBody.SenderPaymentDetails,
		requestBody.ReceiverPaymentDetails,
	)

	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to fetch account numbers")
		return err
	}

	ref := client.Collection("transaction")
	docRef, _, err := ref.Add(ctx, transaction)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to store transaction in Firestore")
		return err
	}

	_, errUpdate := ref.Doc(docRef.ID).Update(ctx, []firestore.Update{
		{Path: "ID", Value: docRef.ID},
	})
	if errUpdate != nil {
		log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction ID")
		return errUpdate
	}

	if requestBody.ReceiverID != "" {
		_, errUpdate = ref.Doc(docRef.ID).Update(ctx, []firestore.Update{
			{Path: "ReceiverID", Value: requestBody.ReceiverID},
		})
		if errUpdate != nil {
			log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
			return errUpdate
		}
	}

	if err := processTransaction(ctx, client, senderAccNo, receiverAccNo, requestBody.Amount, strings.ToUpper(requestBody.PaymentMethod)); err != nil {
		log.Logger.Error().Err(err).Msg("Payment processing failed, updating status to failed")

		_, errUpdate := ref.Doc(docRef.ID).Update(ctx, []firestore.Update{
			{Path: "Status", Value: "fail"},
		})
		if errUpdate != nil {
			log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
		}
		return err
	}

	_, errUpdate = ref.Doc(docRef.ID).Update(ctx, []firestore.Update{
		{Path: "Status", Value: "success"},
	})
	if errUpdate != nil {
		log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
		return errUpdate
	}

	return nil
}

func processTransaction(ctx context.Context, client *firestore.Client, senderAccNo, recipientAccNo string, amount float64, paymentMethod string) error {

	MapPaymentAmount, err := config.GetPaymentAmountYamlConfig()
	if err != nil {
		return fmt.Errorf("unable to load payment configuration: %w", err)
	}

	switch paymentMethod {
	case "UPI":
		if amount > MapPaymentAmount.MaxUpiAmount {
			log.Logger.Error().
				Str("paymentMethod", paymentMethod).
				Float64("amount", amount).
				Msg("UPI payment amount exceeds the maximum allowed limit")
			return fmt.Errorf("UPI payment amount exceeds the maximum allowed limit of %v", MapPaymentAmount.MaxUpiAmount)
		}
	case "CREDIT":
		if amount > MapPaymentAmount.MaxCreditAmount {
			log.Logger.Error().
				Str("paymentMethod", paymentMethod).
				Float64("amount", amount).
				Float64("maxAmount", MapPaymentAmount.MaxCreditAmount).
				Msg("Credit card payment amount exceeds the maximum allowed limit")
			return fmt.Errorf("Credit card payment amount exceeds the maximum allowed limit of %v", MapPaymentAmount.MaxCreditAmount)
		}
	default:
		log.Logger.Error().
			Str("paymentMethod", paymentMethod).
			Msg("Invalid payment method")
		return fmt.Errorf("invalid payment method: %s", paymentMethod)
	}

	bankDetailsRef := client.Collection("BankDetails")

	senderQuery := bankDetailsRef.Where("account_number", "==", senderAccNo).Documents(ctx)
	senderDoc, err := senderQuery.Next()
	if err != nil {
		if err == iterator.Done {
			log.Logger.Error().
				Str("senderAccNo", senderAccNo).
				Msg("No matching document found for sender")
			return fmt.Errorf("no matching document found for sender with account number: %s", senderAccNo)
		}
		log.Logger.Error().Err(err).Msg("Error fetching sender document")
		return fmt.Errorf("failed to fetch sender document: %v", err)
	}

	senderData := senderDoc.Data()
	senderBalance := senderData["balance"].(float64)
	senderDocRef := senderDoc.Ref

	if senderBalance < amount {
		log.Logger.Error().
			Float64("balance", senderBalance).
			Float64("amount", amount).
			Msg("Insufficient balance in sender's account")
		return fmt.Errorf("insufficient balance in sender's account")
	}

	newSenderBalance := senderBalance - amount
	_, err = senderDocRef.Update(ctx, []firestore.Update{
		{Path: "balance", Value: newSenderBalance},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to update sender's balance")
		return fmt.Errorf("failed to update sender's balance: %v", err)
	}

	receiverQuery := bankDetailsRef.Where("account_number", "==", recipientAccNo).Documents(ctx)
	receiverDoc, err := receiverQuery.Next()
	if err != nil {
		if err == iterator.Done {
			log.Logger.Error().
				Str("recipientAccNo", recipientAccNo).
				Msg("No matching document found for receiver")
			return fmt.Errorf("no matching document found for receiver with account number: %s", recipientAccNo)
		}
		log.Logger.Error().Err(err).Msg("Error fetching receiver document")
		return fmt.Errorf("failed to fetch receiver document: %v", err)
	}

	receiverData := receiverDoc.Data()
	receiverBalance := receiverData["balance"].(float64)
	receiverDocRef := receiverDoc.Ref

	newReceiverBalance := receiverBalance + amount
	_, err = receiverDocRef.Update(ctx, []firestore.Update{
		{Path: "balance", Value: newReceiverBalance},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to update receiver's balance")
		return fmt.Errorf("failed to update receiver's balance: %v", err)
	}

	return nil
}

func MakeRequest(ctx context.Context, requestBody entity.MakePaymentRequest) error {
	transaction := transactionPool.Get().(*entity.Transaction)
	defer func() {
		resetTransaction(transaction)
		transactionPool.Put(transaction)
	}()

	transaction.SenderID = requestBody.PayerID
	transaction.ReceiverID = requestBody.RequesterID
	transaction.Amount = requestBody.Amount
	transaction.PaymentMethod = strings.ToUpper(requestBody.PayerPaymentMethod)
	transaction.RecievingMethod = strings.ToUpper(requestBody.RequesterPaymentMethod)
	transaction.SenderPaymentDetails = requestBody.PayerPaymentDetails
	transaction.RecieverPaymentDetails = requestBody.RequesterPaymentDetails
	transaction.TransactionType = "Request"
	transaction.Status = "pending"
	transaction.Timestamp = time.Now().Unix()

	app, err := config.InitFirebase()
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to initialize Firebase app")
		return err
	}

	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get Firestore client")
		return err
	}
	defer client.Close()

	senderAccNo, receiverAccNo, err := repository.GetUserAccNo(ctx, client,
		strings.ToUpper(requestBody.PayerPaymentMethod),
		strings.ToUpper(requestBody.RequesterPaymentMethod),
		requestBody.PayerPaymentDetails,
		requestBody.RequesterPaymentDetails,
	)

	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to fetch account numbers")
		return err
	}

	ref := client.Collection("transaction")
	docRef, _, err := ref.Add(ctx, transaction)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to store transaction in Firestore")
		return err
	}

	_, errUpdate := ref.Doc(docRef.ID).Update(ctx, []firestore.Update{
		{Path: "ID", Value: docRef.ID},
	})
	if errUpdate != nil {
		log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction ID")
		return errUpdate
	}

	requestTransaction := transactionRequestPool.Get().(*entity.TransactionRequest)
	defer func() {
		resetTransactionRequest(requestTransaction)
		transactionRequestPool.Put(requestTransaction)
	}()

	requestTransaction.RequesterAccNo = senderAccNo
	requestTransaction.PayerAccNo = receiverAccNo
	requestTransaction.Amount = requestBody.Amount
	requestTransaction.PaymentMethod = requestBody.RequesterPaymentMethod
	requestTransaction.TransactionID = docRef.ID
	requestTransaction.From = requestBody.RequesterID
	requestTransaction.To = requestBody.PayerID

	requestRef := client.Collection("TransactionRequest")
	requestDocRef, _, err := requestRef.Add(ctx, requestTransaction)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to store transaction in Firestore")
		return err
	}

	_, errUpdate = requestRef.Doc(requestDocRef.ID).Update(ctx, []firestore.Update{
		{Path: "ID", Value: requestDocRef.ID},
	})
	if errUpdate != nil {
		log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction ID")
		return errUpdate
	}

	return nil

}

var transactionLocks sync.Map

func GetTransactionLock(transactionID string) *sync.Mutex {
	lock, exists := transactionLocks.Load(transactionID)
	if !exists {
		lock = &sync.Mutex{}
		transactionLocks.Store(transactionID, lock)
	}
	return lock.(*sync.Mutex)
}

func PaymentRequestAction(ctx context.Context, requestBody entity.PaymentRequestAction, userID string) error {

	transactionLock := GetTransactionLock(requestBody.RequestID)
	transactionLock.Lock() // Acquire the lock

	defer transactionLock.Unlock() // Ensure the lock is released when the function completes

	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Firebase app")
		return err
	}

	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Firestore client")
		return err
	}
	defer client.Close()

	transactionRequestDetails := client.Collection("TransactionRequest")
	requestDocRef := transactionRequestDetails.Doc(requestBody.RequestID)
	requestDoc, err := requestDocRef.Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = " {
			log.Logger.Error().Msg("No matching document found for Request ID")
			return fmt.Errorf("no matching document found for Request ID : %s", requestBody.RequestID)
		}
		log.Logger.Error().Err(err).Msg("Error fetching request document")
		return fmt.Errorf("failed to fetch sender document: %v", err)
	}

	requestData := requestDoc.Data()

	transactionRef := client.Collection("transaction")
	transactionDocRef := transactionRef.Doc(requestData["TransactionID"].(string))
	transactionDoc, err := transactionDocRef.Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = " {
			log.Logger.Error().Msg("No matching document found for Transaction ID")
			return fmt.Errorf("no matching document found for Transaction ID : %s", requestData["TransactionID"].(string))
		}
		log.Logger.Error().Err(err).Msg("Error fetching transaction document")
		return fmt.Errorf("failed to fetch transaction document: %v", err)
	}

	transactionData := transactionDoc.Data()

	if strings.EqualFold(requestBody.Action, "Accept") {
		// Check if the payer is the same as the requester and ensure they are the user attempting the action
		if strings.EqualFold(requestData["To"].(string), transactionData["SenderID"].(string)) && strings.EqualFold(requestData["To"].(string), userID) {

			if err := processTransaction(ctx, client, requestData["PayerAccNo"].(string), requestData["RequesterAccNo"].(string), requestData["Amount"].(float64), strings.ToUpper(requestData["PaymentMethod"].(string))); err != nil {
				log.Logger.Error().Err(err).Msg("Payment processing failed, updating status to failed")

				_, errUpdate := transactionRef.Doc(transactionDocRef.ID).Update(ctx, []firestore.Update{
					{Path: "Status", Value: "fail"},
				})
				if errUpdate != nil {
					log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
				}
				return err
			}

			_, errUpdate := transactionRef.Doc(transactionDocRef.ID).Update(ctx, []firestore.Update{
				{Path: "Status", Value: "success"},
			})
			if errUpdate != nil {
				log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
				return errUpdate
			}
		} else {
			return fmt.Errorf("Invalid User")
		}
	} else if strings.EqualFold(requestBody.Action, "Cancel") {
		if (strings.EqualFold(userID, requestData["From"].(string)) && strings.EqualFold(userID, transactionData["ReceiverID"].(string))) || (strings.EqualFold(userID, requestData["To"].(string)) && strings.EqualFold(userID, transactionData["SenderID"].(string))) {

			_, errUpdate := transactionRef.Doc(transactionDocRef.ID).Update(ctx, []firestore.Update{
				{Path: "Status", Value: "cancel"},
			})
			if errUpdate != nil {
				log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
				return errUpdate
			}
		} else {
			return fmt.Errorf("Invalid User")
		}
	} else {
		_, errUpdate := transactionRef.Doc(transactionDocRef.ID).Update(ctx, []firestore.Update{
			{Path: "Status", Value: "fail"},
		})
		if errUpdate != nil {
			log.Logger.Error().Err(errUpdate).Msg("Failed to update transaction status")
			return errUpdate
		}
	}

	_, errUpdate := transactionRef.Doc(transactionDocRef.ID).Update(ctx, []firestore.Update{
		{Path: "ActionBy", Value: userID},
	})
	if errUpdate != nil {
		return fmt.Errorf("Failed to update transaction ActionBy : %v", errUpdate)
	}

	_, err = requestDocRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete request document: %v", err)
	}

	return nil
}
func GetTransactionByID(ctx context.Context, docID, role, userID string) (*entity.Transaction, error) {
	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Firebase app")
		return nil, err
	}

	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Firestore client")
		return nil, err
	}
	defer client.Close()

	docRef := client.Collection("transaction").Doc(docID)

	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = " {
			log.Error().
				Str("document_id", docID).
				Msg("Transaction document not found")
			return nil, fmt.Errorf("no transaction found with document ID: %s", docID)
		}
		log.Error().Err(err).Msg("Error fetching transaction document")
		return nil, fmt.Errorf("failed to fetch transaction document: %v", err)
	}

	var transaction entity.Transaction
	err = docSnap.DataTo(&transaction)
	if err != nil {
		log.Error().Err(err).Msg("Failed to map Firestore document to struct")
		return nil, fmt.Errorf("failed to map Firestore document: %v", err)
	}

	if strings.EqualFold(role, "ADMIN") || (strings.EqualFold(role, "USER") && (strings.EqualFold(transaction.SenderID, userID) || strings.EqualFold(transaction.ReceiverID, userID))) {

	} else {
		return nil, fmt.Errorf("Invalid User : %s %s", userID, role)
	}

	transaction.ID = docSnap.Ref.ID

	return &transaction, nil
}

func GetTransactions(ctx context.Context, role, userID string, pageSize, pageNumber int) ([]*entity.Transaction, error) {
	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Firebase app")
		return nil, err
	}

	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Firestore client")
		return nil, err
	}
	defer client.Close()

	transactionsRef := client.Collection("transaction")

	var transactions []*entity.Transaction
	transactionMap := make(map[string]*entity.Transaction)

	if strings.ToUpper(role) == "ADMIN" {
		iter := transactionsRef.Documents(ctx)
		defer iter.Stop()

		for {
			docSnap, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Error().Err(err).Msg("Error fetching transactions")
				return nil, fmt.Errorf("failed to fetch transactions: %v", err)
			}

			var transaction entity.Transaction
			err = docSnap.DataTo(&transaction)
			if err != nil {
				log.Error().Err(err).Msg("Failed to map Firestore document to struct")
				return nil, fmt.Errorf("failed to map Firestore document: %v", err)
			}

			transaction.ID = docSnap.Ref.ID
			transactions = append(transactions, &transaction)
		}
	} else if strings.ToUpper(role) == "USER" {
		senderIter := transactionsRef.Where("SenderID", "==", userID).Documents(ctx)
		defer senderIter.Stop()

		for {
			docSnap, err := senderIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Error().Err(err).Msg("Error fetching sender transactions")
				return nil, fmt.Errorf("failed to fetch sender transactions: %v", err)
			}

			var transaction entity.Transaction
			err = docSnap.DataTo(&transaction)
			if err != nil {
				log.Error().Err(err).Msg("Failed to map sender Firestore document to struct")
				return nil, fmt.Errorf("failed to map sender Firestore document: %v", err)
			}

			transaction.ID = docSnap.Ref.ID
			transactionMap[transaction.ID] = &transaction
		}

		receiverIter := transactionsRef.Where("ReceiverID", "==", userID).Documents(ctx)
		defer receiverIter.Stop()

		for {
			docSnap, err := receiverIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Error().Err(err).Msg("Error fetching receiver transactions")
				return nil, fmt.Errorf("failed to fetch receiver transactions: %v", err)
			}

			var transaction entity.Transaction
			err = docSnap.DataTo(&transaction)
			if err != nil {
				log.Error().Err(err).Msg("Failed to map receiver Firestore document to struct")
				return nil, fmt.Errorf("failed to map receiver Firestore document: %v", err)
			}

			transaction.ID = docSnap.Ref.ID
			transactionMap[transaction.ID] = &transaction
		}

		for _, transaction := range transactionMap {
			transactions = append(transactions, transaction)
		}
	} else {
		return nil, fmt.Errorf("invalid role: %v", strings.ToUpper(role))
	}

	startIndex := (pageNumber - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex > len(transactions) {
		return []*entity.Transaction{}, nil
	}
	if endIndex > len(transactions) {
		endIndex = len(transactions)
	}

	return transactions[startIndex:endIndex], nil
}
