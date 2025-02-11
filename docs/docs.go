package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "schemes": {{ marshal .Schemes }},
    "paths": {
        "/initiate": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Transaction APIs"
                ],
                "summary": "Transaction initiate payment.",
                "parameters": [
                    {
                        "description": "Transaction Request Data",
                        "name": "transactionData",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/request.InitiateTransactionPaymentRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    },
                    "500": {
                        "description": "Server Failure"
                    }
                }
            }
        },
        "/login": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Transaction APIs"
                ],
                "summary": "Login user",
                "description": "Logs a user in with the provided credentials and returns a JWT token if successful.",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "description": "User login credentials",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/request.Login"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Login successful",
                        "schema": {
                            "$ref": "#/definitions/entity.SuccessfulResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid login credentials",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    }
                }
            }
        },
        "/txnID/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Transaction APIs"
                ],
                "summary": "Get Transaction by ID",
                "description": "Fetches transaction details by the provided transaction ID.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "required": true,
                        "type": "string",
                        "description": "Transaction ID"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Transaction details",
                        "schema": {
                            "$ref": "#/definitions/entity.TransactionResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid transaction ID",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    },
                    "500": {
                        "description": "Server Failure",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    }
                }
            }
        },
        "/txnID": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Transaction APIs"
                ],
                "summary": "Get Transaction by ID without parameter",
                "description": "Fetches transaction details when no specific transaction ID is provided.",
                "parameters": [],
                "responses": {
                    "200": {
                        "description": "Transaction details",
                        "schema": {
                            "$ref": "#/definitions/entity.TransactionResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid request",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    },
                    "500": {
                        "description": "Server Failure",
                        "schema": {
                            "$ref": "#/definitions/entity.CommonResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "entity.CommonResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "integer"
                },
                "msg": {
                    "type": "string"
                }
            }
        },
        "entity.SuccessfulResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "Login successful"
                },
                "token": {
                    "type": "string",
                    "example": "jwt_token_here"
                }
            }
        },
        "entity.TransactionResponse": {
            "type": "object",
            "properties": {
                "transaction": {
                    "type": "object",
                    "description": "Transaction data object"
                }
            }
        },
        "request.Login": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "example": "user123"
                },
                "email": {
                    "type": "string",
                    "example": "user@example.com"
                },
                "password": {
                    "type": "string",
                    "example": "password123",
                    "required": true
                }
            }
        },
        "request.InitiateTransactionPaymentRequest": {
            "type": "object",
            "properties": {
                "sender_id": {
                    "type": "string"
                },
                "amount": {
                    "type": "number",
                    "format": "float"
                },
                "payment_method": {
                    "type": "string"
                },
                "recieving_method": {
                    "type": "string"
                },
                "sender_payment_details": {
                    "$ref": "#/definitions/request.PaymentDetails"
                },
                "receiver_payment_details": {
                    "$ref": "#/definitions/request.PaymentDetails"
                }
            }
        },
        "request.PaymentDetails": {
            "type": "object",
            "properties": {
                "upi": {
                    "$ref": "#/definitions/request.UPIDetails"
                },
                "credit_card": {
                    "$ref": "#/definitions/request.CreditCardDetails"
                },
                "bank_details": {
                    "$ref": "#/definitions/request.BankDetails"
                }
            }
        },
        "request.UPIDetails": {
            "type": "object",
            "properties": {
                "upi_id": {
                    "type": "string"
                }
            }
        },
        "request.CreditCardDetails": {
            "type": "object",
            "properties": {
                "card_id": {
                    "type": "string"
                },
                "card_number": {
                    "type": "string"
                }
            }
        },
        "request.BankDetails": {
            "type": "object",
            "properties": {
                "account_number": {
                    "type": "string"
                },
                "ifsc_code": {
                    "type": "string"
                },
                "bank_name": {
                    "type": "string"
                }
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
