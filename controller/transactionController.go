package controller

import (
	"context"
	"encoding/json"
	"go-transaction/entity"
	"go-transaction/service"
	"go-transaction/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func InitiateTransaction(c *gin.Context) {
	var requestBody entity.RequestBody
	var responseBody entity.CommonResponse

	err := utils.ReadRequestBody(c.Request, &requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing request body")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusBadRequest, responseBody)
		return
	}

	ctx := context.Background()

	err = service.InitiateTransaction(ctx, requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error processing transaction")
		responseBody.ApplyResponseBody(entity.FAILURE)

		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	responseBody.ApplyResponseBody(entity.SUCCESS)
	c.JSON(http.StatusOK, responseBody)
}

func MakeRequest(c *gin.Context) {
	var requestBody entity.MakePaymentRequest
	var responseBody entity.CommonResponse

	err := utils.ReadMakePaymentRequest(c.Request, &requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing request body")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusBadRequest, responseBody)
		return
	}

	ctx := context.Background()

	err = service.MakeRequest(ctx, requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error processing transaction")
		responseBody.ApplyResponseBody(entity.FAILURE)

		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	responseBody.ApplyResponseBody(entity.SUCCESS)
	c.JSON(http.StatusOK, responseBody)
}

func PaymentRequestAction(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")
	token := authHeader[len("Bearer "):]
	payload, e := utils.GetPayloadFromJWT(token)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Error fetching Token")
		return
	}

	var responseBody entity.CommonResponse
	var requestBody entity.PaymentRequestAction

	uid, ok := payload["uid"].(string)
	if !ok {
		log.Error().
			Msg("Error processing payload")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	err := utils.ReadPaymentRequestAction(c.Request, &requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing request body")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusBadRequest, responseBody)
		return
	}

	ctx := context.Background()

	err = service.PaymentRequestAction(ctx, requestBody, uid)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error processing transaction")
		responseBody.ApplyResponseBody(entity.FAILURE)

		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	responseBody.ApplyResponseBody(entity.SUCCESS)
	c.JSON(http.StatusOK, responseBody)
}

func GetTransactionByID(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := authHeader[len("Bearer "):]
	payload, e := utils.GetPayloadFromJWT(token)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Error fetching Token")
		return
	}

	id := c.Param("id")
	var responseBody entity.CommonResponse
	ctx := context.Background()

	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber"))

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNumber <= 0 {
		pageNumber = 1
	}

	var transaction interface{}
	var err error

	uid, ok := payload["uid"].(string)
	if !ok {
		log.Error().
			Msg("Error processing payload")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	role, ok := payload["role"].(string)
	if !ok {
		log.Error().
			Msg("Error processing payload")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	if id == "" {
		transaction, err = service.GetTransactions(ctx, role, uid, pageSize, pageNumber)
	} else {
		transaction, err = service.GetTransactionByID(ctx, id, role, uid)
	}

	if err != nil {
		log.Error().
			Err(err).
			Msg("Error fetching transaction ID ")
		responseBody.ApplyResponseBody(entity.FAILURE)
		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	responseBody.ApplyResponseBody(entity.SUCCESS)
	response := gin.H{
		"data": transaction,
		"metadata": gin.H{
			"status": responseBody,
		},
	}

	if id == "" {
		transactions, ok := transaction.([]*entity.Transaction)
		if !ok {
			log.Error().Msg("Failed to cast transaction to []*entity.Transaction")
			responseBody.ApplyResponseBody(entity.COMMON_SERVER_ERROR)
			c.JSON(http.StatusInternalServerError, responseBody)
			return
		}

		response["metadata"].(gin.H)["pageSize"] = pageSize
		response["metadata"].(gin.H)["pageNumber"] = pageNumber
		response["metadata"].(gin.H)["totalCount"] = len(transactions)
		if strings.ToUpper(role) == "USER" {
			var received, sent int = 0, 0

			for _, transaction := range transactions {
				if transaction.ReceiverID == uid {
					received++
				}
				if transaction.SenderID == uid {
					sent++
				}
			}

			response["metadata"].(gin.H)["sentCount"] = sent
			response["metadata"].(gin.H)["receivedCount"] = received
		}
	}

	// c.JSON(http.StatusOK, response)
	responseData, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response to JSON")
		responseBody.ApplyResponseBody(entity.COMMON_SERVER_ERROR)
		c.JSON(http.StatusInternalServerError, responseBody)
		return
	}

	// Send the marshalled JSON data as response
	c.Data(http.StatusOK, "application/json", responseData)
}
