package main

import (
	"fmt"

	"go-transaction/config"
	"go-transaction/docs"
	"go-transaction/routes"

	"github.com/rs/zerolog/log"
	
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
