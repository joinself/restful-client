package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joho/godotenv"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/qiangxue/go-env"
)

const (
	defaultServerPort                    = 8080
	defaultJWTExpirationHours            = 72
	defaultRefreshTokenExpirationInHours = 128
	defaultCleanupPeriod                 = 15 // 15 days
)

// Self config object
type SelfAppConfig struct {
	SelfAppID           string `yaml:"self_app_id"`
	SelfAppDeviceSecret string `yaml:"self_device_secret"`
	SelfStorageKey      string `yaml:"self_storage_key"`
	SelfStorageDir      string `yaml:"self_storage_dir"`
	SelfEnv             string `yaml:"self_env"`
	SelfAPIURL          string `yaml:"self_api_url"`
	SelfMessagingURL    string `yaml:"self_messaging_url"`
	CallbackURL         string `yaml:"message_notification_url"`
	DLCode              string `yaml:"dl_code" env:"DL_CODE"`
}

// Config represents an application configuration.
type Config struct {
	// File Filesystem YAML based configuration.
	DefaultSelfApp *SelfAppConfig

	// OPTIONAL based configuration
	// RefreshTokenExpiration JWT refresh expiration in hours.
	RefreshTokenExpirationInHours int `env:"REFRESH_TOKEN_EXPIRATION"`
	// JWTExpiration JWT expiration in hours.
	JWTExpirationTimeInHours int `env:"JWT_EXPIRATION_TIME_IN_HOURS"`
	// ServeDocs string _(true|false)_ defining if docs should be served from the localhost on "/docs" path.
	ServeDocs string `env:"SERVE_DOCS"`
	// ServerPort the server port. Defaults to 8080
	ServerPort int `env:"SERVER_PORT"`
	// DefaultAppCallbackURL the default callback url for any incoming messages.
	DefaultAppCallbackURL string `env:"APP_MESSAGE_NOTIFICATION_URL"`

	// REQUIRED ENV based configuration
	// DSN database data source name
	DSN string `env:"DSN"`
	// JWTSigningKey The signing key used to build the jwt tokens shared with the api clients.
	JWTSigningKey string `env:"JWT_SIGNING_KEY"`
	// User The default user used on the authentication endpoints.
	User string `env:"USER"`
	// Password The default password used on the authentication endpoints.
	Password string `env:"PASSWORD"`
	// StorageDir The default user used on the authentication endpoints.
	StorageDir string `env:"STORAGE_DIR"`
	// StorageKey The default storage key used to encrypt sessions.
	StorageKey string `env:"STORAGE_KEY"`
	// DefaultAppID the default self app identifier.
	DefaultAppID string `env:"APP_ID"`
	// DefaultAppSecret the default self app secret.
	DefaultAppSecret string `env:"APP_SECRET"`
	// DefaultAppEnv the default self app environment.
	DefaultAppEnv string `env:"APP_ENV"`
	// CleanupPeriod the number of days the database temporary data will be removed.
	CleanupPeriod int `env:"CLEANUP_PERIOD"`
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.JWTSigningKey, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.StorageDir, validation.Required),
		validation.Field(&c.StorageKey, validation.Required),
		validation.Field(&c.DSN, validation.Required),
		// validation.Field(&c.DefaultAppID, validation.Required), // this is no longer required
		// validation.Field(&c.DefaultAppEnv, validation.Required), // this is no longer required
	)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(logger log.Logger, envPath string) (*Config, error) {
	erro := godotenv.Load(envPath)
	if erro != nil {
		logger.Debug(".env file not found, using system environment instead")
	}

	// default config
	c := Config{
		RefreshTokenExpirationInHours: defaultRefreshTokenExpirationInHours,
		JWTExpirationTimeInHours:      defaultJWTExpirationHours,
		ServeDocs:                     "false",
		ServerPort:                    defaultServerPort,
		CleanupPeriod:                 defaultCleanupPeriod,
	}

	// load from environment variables prefixed with "APP_"
	if err := env.New("RESTFUL_CLIENT_", logger.Debugf).Load(&c); err != nil {
		return nil, err
	}

	if c.DefaultAppID != "" {
		c.DefaultSelfApp = &SelfAppConfig{
			SelfAppID:           c.DefaultAppID,
			SelfAppDeviceSecret: c.DefaultAppSecret,
			SelfStorageKey:      c.StorageKey,
			SelfStorageDir:      c.StorageDir,
			SelfEnv:             c.DefaultAppEnv,
			CallbackURL:         c.DefaultAppCallbackURL,
		}
	}

	// validation
	return &c, c.Validate()
}
