package config

import (
	"io/ioutil"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v2"
)

const (
	defaultServerPort                    = 8080
	defaultJWTExpirationHours            = 72
	defaultRefreshTokenExpirationInHours = 128
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
	SelfApps []SelfAppConfig `yaml:"self_apps"`

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
	ClientConfigFile      string `env:"CONFIG_FILE"`

	// REQUIRED ENV based configuration
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
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.JWTSigningKey, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.StorageDir, validation.Required),
		validation.Field(&c.StorageKey, validation.Required),
		validation.Field(&c.DefaultAppID, validation.Required),
		validation.Field(&c.DefaultAppEnv, validation.Required),
	)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		RefreshTokenExpirationInHours: defaultRefreshTokenExpirationInHours,
		JWTExpirationTimeInHours:      defaultJWTExpirationHours,
		ServeDocs:                     "false",
		ServerPort:                    defaultServerPort,
	}

	// load from environment variables prefixed with "APP_"
	if err := env.New("RESTFUL_CLIENT_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	if c.DefaultAppID != "" {
		defaultApp := SelfAppConfig{
			SelfAppID:           c.DefaultAppID,
			SelfAppDeviceSecret: c.DefaultAppSecret,
			SelfStorageKey:      c.StorageKey,
			SelfStorageDir:      c.StorageDir,
			SelfEnv:             c.DefaultAppEnv,
			CallbackURL:         c.DefaultAppCallbackURL,
		}
		c.SelfApps = []SelfAppConfig{defaultApp}
	}

	if c.ClientConfigFile != "" {
		// Load extra apps
		// TODO: Load extra apps from the provided yaml config.
		bytes, err := ioutil.ReadFile(c.ClientConfigFile)
		if err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(bytes, &c); err != nil {
			return nil, err
		}

	}

	// validation
	return &c, c.Validate()
}
