package config

import (
	"context"
	"fmt"
	"go-transaction/entity"

	"github.com/rs/zerolog/log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/subosito/gotenv"
)

// InitFirebase initializes the Firebase application with the given credentials and project ID.
// It returns the Firebase app instance or an error if the initialization fails.
func InitFirebase() (*firebase.App, error) {
	filepath := "/etc/render/secrets/firebase.json"
	projectID := "crud-b5a48"
	opt := option.WithCredentialsFile(filepath)

	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: projectID,
	}, opt)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing Firebase app")
		return nil, err
	}
	return app, nil
}

// GetFirestoreClient returns a Firestore client instance for the Firebase app.
// If the client cannot be created, it returns an error.
func GetFirestoreClient(app *firebase.App) (*firestore.Client, error) {
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error getting Firestore client")
		return nil, err
	}
	return client, nil
}

func FirebaseInitialization() (*firestore.Client,error){
	app, err := InitFirebase()
	if err != nil {
		return nil, err
	}

	// Get Firestore client
	client, err := GetFirestoreClient(app)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client, nil
}

// GetServerYamlConfig loads and returns the server configuration from the YAML file.
// It uses the project environment loaded from .env file and reads the config file accordingly.
// It returns a ServerConfig struct or an error if the file cannot be read or unmarshaled.
func GetServerYamlConfig() (*entity.ServerConfig, error) {
	var path = fmt.Sprintf("D:/GO Project/go-transaction/config/config.%s.yaml", ReadEnvConfig())

	var serverConfig entity.ServerConfig

	y := koanf.New(".")
	err := y.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		log.Error().Err(err).Msg("Error reading server config YAML")
		return nil, fmt.Errorf("unable to read config: %v", err)
	}

	err = y.Unmarshal("", &serverConfig)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling server config")
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	return &serverConfig, nil
}

// GetPaymentAmountYamlConfig loads and returns the payment configuration from the YAML file.
// It reads the configuration based on the project environment and unmarshals it into a PaymentConfig struct.
func GetPaymentAmountYamlConfig() (*entity.PaymentConfig, error) {
	var path = fmt.Sprintf("D:/GO Project/go-transaction/config/config.%s.yaml", ReadEnvConfig())

	var paymentConfig entity.PaymentConfig

	y := koanf.New(".")
	err := y.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		log.Error().Err(err).Msg("Error reading payment config YAML")
		return nil, fmt.Errorf("unable to read config: %v", err)
	}

	err = y.Unmarshal("paymentconfig", &paymentConfig)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling payment config")
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	return &paymentConfig, nil
}

// GetSwaggerYamlConfig loads and returns the Swagger configuration from the YAML file.
// It reads the Swagger section of the configuration and unmarshals it into a Swagger struct.
func GetSwaggerYamlConfig() (*entity.Swagger, error) {
	var path = fmt.Sprintf("D:/GO Project/go-transaction/config/config.%s.yaml", ReadEnvConfig())

	var swagger entity.Swagger

	k := koanf.New(".")
	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		log.Error().Err(err).Msg("Error reading Swagger config YAML")
		return nil, fmt.Errorf("unable to read config: %v", err)
	}

	err = k.Unmarshal("swagger", &swagger)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling Swagger config")
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	return &swagger, nil
}

// GetApiYamlConfig loads and returns the API configuration from the YAML file.
// It reads the API section of the configuration and unmarshals it into an Api struct.
func GetApiYamlConfig() (*entity.Api, error) {
	var path = fmt.Sprintf("D:/GO Project/go-transaction/config/config.%s.yaml", ReadEnvConfig())

	var api entity.Api

	k := koanf.New(".")
	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		log.Error().Err(err).Msg("Error reading API config YAML")
		return nil, fmt.Errorf("unable to read config: %v", err)
	}

	err = k.Unmarshal("", &api)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling API config")
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	return &api, nil
}

// ReadEnvConfig loads the environment configuration from the .env file.
// It reads the "PROJECT" value from the environment variables and returns it.
func ReadEnvConfig() string {
	if err := gotenv.Load("D:/GO Project/go-transaction/.env"); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	var k = koanf.New(".")
	err := k.Load(env.Provider("", ".", func(s string) string {
		return s
	}), nil)

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading environment variables")
	}

	project := k.String("PROJECT")
	return project
}

// GetKey loads the secret key and secret string from the environment configuration.
// It reads the values of "SECRET_KEY" and "SECRET_STRING" and returns them.
func GetKey() (string, string) {
	if err := gotenv.Load("D:/GO Project/go-transaction/.env"); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	var k = koanf.New(".")
	err := k.Load(env.Provider("", ".", func(s string) string {
		return s
	}), nil)

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading environment variables")
	}

	key := k.String("SECRET_KEY")
	s := k.String("SECRET_STRING")
	return key, s
}
