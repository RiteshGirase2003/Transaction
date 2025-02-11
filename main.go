package main

import (
	"context"
	"fmt"

	"go-transaction/config"
	"go-transaction/docs"
	"go-transaction/routes"

	"github.com/rs/zerolog/log"

	"github.com/rabbitmq/amqp091-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// rabbitMQURL defines the connection string for RabbitMQ.
const rabbitMQURL = "amqp://guest:guest@localhost:5672/"

// publishToRabbitMQ publishes a message to the specified RabbitMQ queue.
//
// It declares the queue if it does not exist and then sends the message.
// If an error occurs during the process, it returns an error.
func publishToRabbitMQ(ch *amqp091.Channel, queueName, body string) error {
	_, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	err = ch.PublishWithContext(
		context.Background(),
		"",        // Exchange
		queueName, // Routing Key
		false,     // Mandatory
		false,     // Immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	

	log.Printf("✅ Transaction published || queue '%s'",  queueName)
	return nil
}

// listenToFirestoreUpdates listens for real-time updates from Firestore.
//
// It establishes a connection with Firestore and RabbitMQ, continuously
// listening for changes in the "transaction" collection and publishing updates to RabbitMQ.
func listenToFirestoreUpdates() {
	// Initialize Firebase
	app, err := config.InitFirebase()
	if err != nil {
		log.Fatal().Err(err).Msg("Firebase initialization failed")
	}

	// Get Firestore client
	client, err := config.GetFirestoreClient(app)
	if err != nil {
		log.Fatal().Err(err).Msg("Firestore client initialization failed")
	}
	defer client.Close()

	lastConsumedRef := client.Collection("LastConsumed").Doc("lastConsume")
	doc, err := lastConsumedRef.Get(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get 'LastConsumed' document")
	}

	// Retrieve the timestamp from the document (assuming it's stored as a Unix timestamp)
	lastConsumedTime, ok := doc.Data()["timestamp"].(int64)
	if !ok {
		log.Fatal().Msg("'timestamp' field not found or has incorrect type in 'LastConsumed'")
	}

	// Start listening to Firestore updates
	collectionRef := client.Collection("transaction").Where("Timestamp", ">=", lastConsumedTime)
	iter := collectionRef.Snapshots(context.Background())

	// Establish connection with RabbitMQ
	conn, err := amqp091.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("RabbitMQ connection failed")
	}
	defer conn.Close()

	// Open a channel to communicate with RabbitMQ
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open RabbitMQ channel")
	}
	defer ch.Close()

	// Infinite loop to process Firestore document changes
	for {
		snapshot, err := iter.Next()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Firestore snapshot")
		}

		// Process document changes and send them to RabbitMQ
		for _, docChange := range snapshot.Changes {
			docData := docChange.Doc.Data()
			body := fmt.Sprintf("%v", docData)

			if err := publishToRabbitMQ(ch, "transaction_messages", body); err != nil {
				log.Error().Err(err).Msg("Failed to publish to RabbitMQ")
			}
		}
	}
}

// main initializes Firebase, Firestore, Swagger, and starts the API server.
//
// - Initializes Firebase and Firestore clients.
// - Loads Swagger configuration.
// - Starts listening to Firestore updates in a separate goroutine.
// - Initializes routes and runs the HTTP server.
func main() {
	// Initialize Firebase
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

	// Load Swagger configuration
	swagger, err := config.GetSwaggerYamlConfig()
	if err != nil {
		log.Error().Err(err).Msg("Error loading Swagger YAML configuration")
		return
	}
	docs.SwaggerInfo.Host = swagger.Host
	docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", swagger.BasePath)

	// Start listening to Firestore updates asynchronously
	go listenToFirestoreUpdates()

	
	// Initialize API routes
	router := routes.InitRoutes()
	router.GET(fmt.Sprintf("%s/*any", swagger.Url), ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Load server configuration
	serverConfig, err := config.GetServerYamlConfig()
	if err != nil {
		log.Error().Err(err).Msg("Error loading server YAML configuration")
		return
	}

	log.Info().Msgf("Swagger UI available at: http://localhost:%d/%s/index.html", serverConfig.Port, swagger.Url)

	// Start the HTTP server
	router.Run(fmt.Sprintf(":%d", serverConfig.Port))
}
