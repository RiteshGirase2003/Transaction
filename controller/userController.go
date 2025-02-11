package controller

import (
	"context"
	"go-transaction/entity"
	"go-transaction/service"
	"go-transaction/utils"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

// Login handles the user login process.
// It validates the login credentials from the request body, authenticates the user,
// and generates an authentication token.
//   - If the credentials are invalid, it returns a `400 Bad Request` error response with the error details.
//   - If the authentication fails, an error is logged, and a failure response is sent to the client.
//   - If the login is successful, a JWT token is generated and returned along with a success message.
func Login(c *gin.Context) {
	var credentials entity.Login
	var responseBody entity.CommonResponse

	// Parse the incoming JSON request to extract user credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Create a context for the login process
	ctx := context.Background()

	// Attempt to authenticate the user with the provided credentials
	user, err := service.LoginUser(ctx, credentials)

	if err != nil {
		// Log error and return failure response
		log.Error().
			Err(err).
			Msg("Error")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusBadRequest, responseBody)
		return
	}

	// Generate an authentication token for the authenticated user
	token, err := utils.GenerateAuthToken(user)
	if err != nil {
		// Log error and return failure response
		log.Error().
			Err(err).
			Msg("Error generating token")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusBadRequest, responseBody)
		return
	}

	// Return success response with the generated token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}
