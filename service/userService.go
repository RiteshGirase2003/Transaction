package service

import (
	"context"
	"errors"
	"fmt"
	"go-transaction/config"
	"go-transaction/entity"

	"cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

func LoginUser(ctx context.Context, credentials entity.Login) (*entity.User, error) {
	var user entity.User

	// Initialize Firebase app
	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Firebase app")
		return nil, err
	}

	// Get Firestore client
	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Firestore client")
		return nil, err
	}
	defer client.Close()

	var docRef *firestore.DocumentRef
	var userQuery *firestore.DocumentIterator

	// Determine the query to run based on provided credentials
	if credentials.UserID != "" {
		docRef = client.Collection("users").Doc(credentials.UserID)
	} else if credentials.Email != "" {
		userQuery = client.Collection("users").Where("email", "==", credentials.Email).Limit(1).Documents(ctx)
	} else {
		return nil, errors.New("username or email is required")
	}

	// If using email query, get the first document from the iterator
	if userQuery != nil {
		docSnap, err := userQuery.Next()
		if err == iterator.Done {
			log.Error().Msg("User with the provided email not found")
			return nil, fmt.Errorf("no user found with the provided email")
		} else if err != nil {
			log.Error().Err(err).Msg("Error fetching user document")
			return nil, fmt.Errorf("failed to fetch user document: %v", err)
		}
		// Set docRef to the result from the query
		docRef = docSnap.Ref
	}

	// Fetch user document by document reference
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = " {
			log.Error().Msg("User document not found")
			return nil, fmt.Errorf("no user found")
		}
		log.Error().Err(err).Msg("Error fetching user document")
		return nil, fmt.Errorf("failed to fetch user document: %v", err)
	}

	// Map Firestore document data to User struct
	err = docSnap.DataTo(&user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to map Firestore document to struct")
		return nil, fmt.Errorf("failed to map Firestore document: %v", err)
	}

	// Check if the password matches
	if user.Password != credentials.Password {
		return nil, fmt.Errorf("invalid password")
	}

	// Set document ID manually (since Firestore doesn't store it as part of the document)
	user.UserID = docSnap.Ref.ID

	// Clear the password before returning the user
	user.Password = ""

	// Log success
	log.Info().Str("user_id", user.UserID).Msg("User fetched successfully")

	return &user, nil
}


func GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	var user entity.User

	// Initialize Firebase app
	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Firebase app")
		return nil, err
	}

	// Get Firestore client
	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Firestore client")
		return nil, err
	}
	defer client.Close()

	// Fetch user document by userID
	docRef := client.Collection("users").Doc(userID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = " {
			log.Error().Msg("User document not found")
			return nil, fmt.Errorf("no user found with the provided ID")
		}
		log.Error().Err(err).Msg("Error fetching user document")
		return nil, fmt.Errorf("failed to fetch user document: %v", err)
	}

	// Map Firestore document data to User struct
	err = docSnap.DataTo(&user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to map Firestore document to struct")
		return nil, fmt.Errorf("failed to map Firestore document: %v", err)
	}

	// Set document ID manually (since Firestore doesn't store it as part of the document)
	user.UserID = docSnap.Ref.ID

	// Clear the password before returning the user
	user.Password = ""

	// Log success
	log.Info().Str("user_id", user.UserID).Msg("User fetched successfully")

	return &user, nil
}