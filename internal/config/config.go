package config

import (
	"io/ioutil"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v2"
)

const (
	defaultServerPort         = 8080
	defaultJWTExpirationHours = 72
)

// Self config object
type SelfAppConfig struct {
	SelfAppID           string `yaml:"self_app_id"`
	SelfAppDeviceSecret string `yaml:"self_device_secret"`
	SelfStorageKey      string `yaml:"self_storage_key"`
	SelfStorageDir      string `yaml:"self_storage_dir"`
	SelfEnv             string `yaml:"self_env"`
	CallbackURL         string `yaml:"message_notification_url"`
	DLCode              string `yaml:"dl_code" env:"DL_CODE"`
}

// Config represents an application configuration.
type Config struct {
	// the server port. Defaults to 8080
	ServerPort int `yaml:"server_port" env:"SERVER_PORT"`
	// the data source name (DSN) for connecting to the database. required.
	DSN string `yaml:"dsn" env:"DSN,secret"`
	// JWT signing key. required.
	JWTSigningKey string `yaml:"jwt_signing_key" env:"JWT_SIGNING_KEY,secret"`
	// JWT expiration in hours. Defaults to 72 hours (3 days)
	JWTExpiration          int             `yaml:"jwt_expiration" env:"JWT_EXPIRATION"`
	RefreshTokenExpiration int             `yaml:"refresh_token_expiration" env:"REFRESH_TOKEN_EXPIRATION"`
	SelfApps               []SelfAppConfig `yaml:"self_apps" env:"SELF_APPS"`
	User                   string          `yaml:"user" env:"USER"`
	Password               string          `yaml:"password" env:"PASSWORD"`
	ServeDocs              string          `yaml:"serve_docs" env:"SERVE_DOCS"`
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required),
		validation.Field(&c.JWTSigningKey, validation.Required),
	)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(file string, logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		ServerPort:    defaultServerPort,
		JWTExpiration: defaultJWTExpirationHours,
	}

	// load from YAML config file
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	// load from environment variables prefixed with "APP_"
	if err = env.New("APP_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	// validation
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return &c, err
}
