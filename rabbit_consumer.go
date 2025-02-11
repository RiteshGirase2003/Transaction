package main

import (
	"context"
	"fmt"
	"go-transaction/config"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/googleapi"
)

func rabbit() {
	// Set up zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stderr
		w.NoColor = false
	}))

	app, err := config.InitFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Error initializing Firebase")
		return
	}

	// Get Firestore client
	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Error().Err(err).Msg("Error getting Firestore client")
		return
	}
	defer client.Close()

	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ctx := context.Background()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open a channel")
	}
	defer ch.Close()

	// Declare the queue (must match the queue in publisher)
	queueName := "transaction_messages"
	_, err = ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare queue")
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register consumer")
	}

	fmt.Println("Waiting for messages. To exit press CTRL+C")

	// Firestore collection reference for LastConsumed
	lastConsumedRef := client.Collection("LastConsumed").Doc("lastConsume")

	// Process messages
	for msg := range msgs {
		// Get the document to check if it exists
		doc, err := lastConsumedRef.Get(ctx)
		if err != nil {
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
				// Document does not exist, create it
				_, err := lastConsumedRef.Set(ctx, map[string]interface{}{
					"timestamp": time.Now().Unix(),
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to create 'lastConsume' document")
					return
				}
				log.Info().Msg("Created 'lastConsume' document with timestamp")
			} else {
				// Other error getting the document
				log.Error().Err(err).Msg("Error getting 'lastConsume' document")
				return
			}
		} else {
			// Document exists, update it with the current timestamp
			if doc.Exists() {
				_, err := lastConsumedRef.Update(ctx, []firestore.Update{
					{Path: "timestamp", Value: time.Now().Unix()},
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to update 'lastConsume' document")
					return
				}
				log.Info().Msg("Updated 'lastConsume' document with new timestamp")
			} else {
				log.Warn().Msg("Document 'lastConsume' exists but is empty")
			}
		}

		// Process the message
		log.Info().Msgf("\n\n ✅  %s\n", msg.Body)
	}
}
