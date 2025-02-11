package routes

import (
	"fmt"
	"go-transaction/config"

	"github.com/rs/zerolog/log"

	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
)

// # InitRoutes initializes and returns a configured Gin router instance.
//
// This function:
// 		- Loads the API configuration from a YAML file.
// 		- Applies the gin-zerolog middleware for structured logging.
// 		- Creates a route group based on the API version.
// 		- Delegates the setup of transaction-specific routes to TransactionRoutes().
//
// Returns:
// 		- *gin.Engine: Configured Gin router instance.
// 		- nil if there is an error loading the API configuration.
func InitRoutes() *gin.Engine {
	router := gin.Default()

	api, err := config.GetApiYamlConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load API YAML configuration")
		return nil
	}

	router.Use(ginzerolog.Logger("gin"))

	routerGroup := router.Group(fmt.Sprintf("/%s", api.Api))

	TransactionRoutes(routerGroup)

	return router
}
