// This is entity package
// 
// this 
package entity

// ServerConfig:
// This struct holds the configuration related to the server's settings.
// It contains details such as the port the server listens on.
//
// Fields:
// 	1. Port: The port number the server will listen to for incoming requests.
//
type ServerConfig struct {
	Port int `koanf:"port"`
}

// PaymentConfig:
// This struct contains the configuration for payment-related settings.
// It defines the maximum allowed amounts for different payment methods such as UPI and credit cards.
//
// Fields:
// 	1. MaxUpiAmount: 	The maximum allowed amount for UPI transactions (in float64).
// 	2. MaxCreditAmount: The maximum allowed amount for credit card transactions (in float64).
//
type PaymentConfig struct {
	MaxUpiAmount    float64 `koanf:"upi"`
	MaxCreditAmount float64 `koanf:"credit"`
}

// # Api
//
// This struct holds the configuration related to API settings.
// It typically contains information about API endpoint wrt Production/Staging.
//
// Fields:
// 	1. Api: identifier for the API endpoint (string).
//

type Api struct {
	Api string `koanf:"api"`
}

// Swagger:
// This struct defines the configuration settings for the Swagger API documentation.
// It contains details like the host, base path, and full URL for accessing the Swagger documentation.
//
// Fields:
//
// 	1. Host: 		The host URL for serving the Swagger API documentation (string).
// 	2. BasePath: 	The base path for the Swagger API (string).
// 	3. Url: 		The full URL to access the Swagger API documentation (string).
//
// Example usage:
//

type Swagger struct {
	Host     string `koanf:"host"`
	BasePath string `koanf:"basePath"`
	Url      string `koanf:"url"`
}
