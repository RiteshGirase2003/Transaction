package routes

import (
	"go-transaction/controller"
	"go-transaction/middleware"

	"github.com/gin-gonic/gin"
)

// TransactionRoutes defines the routes related to transaction operations.
//
// These routes include user authentication, transaction initiation, 
// payment request actions, and retrieval of transaction details.
//
// Routes:
//   - POST /login: User authentication endpoint to log in.
//   - POST /initiate: Initiates a transaction, requiring authentication.
//   - POST /make-request: Makes a payment request, requiring authentication.
//   - POST /request-action: Handles actions on payment requests, requiring authentication.
//   - GET /txnID/:id: Retrieves transaction details by transaction ID, requiring authentication.
//   - GET /txnID: Retrieves transaction details, requiring authentication.
func TransactionRoutes(router *gin.RouterGroup) {
	router.POST("/login", controller.Login)
	router.POST("/initiate", middleware.AuthCheck(), controller.InitiateTransaction)
	router.POST("/make-request", middleware.AuthCheck(), controller.MakeRequest)
	router.POST("/request-action", middleware.AuthCheck(), controller.PaymentRequestAction)
	router.GET("/txnID/:id", middleware.AuthCheck(), controller.GetTransactionByID)
	router.GET("/txnID", middleware.AuthCheck(), controller.GetTransactionByID)
}
