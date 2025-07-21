package config

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(score,threshold)
type SearchMode string

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentStaging     = "staging"
)

type (
	// Flags is a struct to enable/disable features
	Flags struct {
		// Debug indicates whether the application is running in debug mode
		Debug bool `mapstructure:"debug"`

		// OpenTelemetry indicates whether OpenTelemetry is enabled
		OpenTelemetry bool `mapstructure:"open_telemetry"`

		// Email indicates whether email services are enabled
		Email bool `mapstructure:"email_services"`
	}

	// SmtpConfig is the configuration for the SMTP server
	SmtpConfig struct {
		// Host is the hostname of the SMTP server (e.g. smtp.gmail.com)
		Host string `mapstructure:"host"`

		// Port is the port of the SMTP server (typically 587 for TLS)
		Port int `mapstructure:"port"`

		// Username is the username for the SMTP server (e.g. no-reply)
		Username string `mapstructure:"username"`

		// Password is the password for the SMTP server
		Password string `mapstructure:"password"`

		// FromName is the name of the sender (e.g. Hubble)
		FromName string `mapstructure:"from_name"`

		// FromUrl is the URL of the sender (e.g. hubble.keystroke.tools)
		FromUrl string `mapstructure:"from_url"`
	}

	LLM struct {
		// BaseURL is the hostname of the OpenAI-compatible server (e.g. localhost:5000)
		BaseURL string `mapstructure:"base_url"`

		// ApiKey is the API key for the OpenAI-compatible server
		ApiKey string `mapstructure:"api_key"`

		// EmbeddingsModel is the model used for embeddings (e.g. text-embedding-ada-002)
		EmbeddingsModel string `mapstructure:"embedding_model"`

		enabledEmbeddings bool `mapstructure:"-"`
	}

	Search struct {
		/*
		   Mode defines the specificity of the search engine

		   By default, it is set to "threshold" which means the search engine will use a threshold-based approach to determine the relevance of the search results.

		   The other options are:
		   - "score": The search engine will only use a score-based approach to determine the relevance of the search results.
		*/
		Mode SearchMode `mapstructure:"mode"`

		// Threshold is the threshold for the search engine (default: 30)
		Threshold float64 `mapstructure:"threshold"`
	}

	// Drivers is the configuration for the modular pieces of the application (e.g. key-value store)
	Drivers struct {
		// The preferred driver for the key-value store
		KV string `mapstructure:"kv"`
	}

	// Keys is the configuration for the encryption keys used by the application for various purposes
	Keys struct {
		// CookieSecret is the secret key used to sign cookies
		CookieSecret string `mapstructure:"cookie_secret"`

		/*
			TotpSecrets should contain all versions of the TOTP secret keys separated by a comma e.g. "v1_secret,v2_secret,v3_secret"

			This is required to enable rolling TOTP secret keys as required, these keys are used to sign the hash stored in the database
		*/
		TotpSecrets string `mapstructure:"totp_secrets"`
	}

	// Minio is the configuration for the object store
	Minio struct {
		// Endpoint is the endpoint of the object store (e.g. s3.amazonaws.com, minio:9000)
		Endpoint string `mapstructure:"endpoint"`

		// AccessKey is the access key for the object store
		AccessKey string `mapstructure:"access_key"`

		// SecretKey is the secret key for the object store
		SecretKey string `mapstructure:"secret_key"`

		// UseSSL indicates whether to use SSL for the object store
		UseSSL bool `mapstructure:"use_ssl"`
	}

	Plugins struct {
		// Directory is the directory where the plugins-related files are stored
		Directory string `mapstructure:"dir"`
	}

	Config struct {
		// Environment is the environment the application is running in (e.g. development, production)
		Environment string `mapstructure:"environment"`

		// Port is the port the server will listen on
		Port int `mapstructure:"port"`

		// Flags to enable/disable features
		Flags Flags `mapstructure:"enable"`

		// AppUrl is the base URL of the application's frontend
		AppUrl string `mapstructure:"app_url"`

		// PostgresDSN is the DSN for the Postgres database
		PostgresDSN string `mapstructure:"postgres_dsn"`

		// BadgerDbPath is the path to the BadgerDB database
		BadgerDbPath string `mapstructure:"badger_db_path"`

		// EtcdEndpoints is the list of etcd endpoints
		EtcdEndpoints []string `mapstructure:"etcd_endpoints"`

		// Smtp is the configuration for the SMTP server
		Smtp SmtpConfig `mapstructure:"smtp"`

		// Drivers is the configuration for the modular pieces of the application
		Drivers Drivers `mapstructure:"driver"`

		// Keys is the configuration for the encryption keys used by the application
		Keys Keys `mapstructure:"key"`

		// Minio is the configuration for the object store
		Minio Minio `mapstructure:"minio"`

		// Plugins is the configuration for the plugins
		Plugins Plugins `mapstructure:"plugins"`

		// LLM is the configuration for the OpenAI-compatible server
		LLM LLM `mapstructure:"llm"`

		// Search is the configuration for the search engine
		Search Search `mapstructure:"search"`

		// TOTP stuff
		totp struct {
			// TOTPKeys is a map of the TOTP secret keys for each version
			keys map[int16]string
			// latestVersion is the pre=computed latest version of the TOTP secret keys
			latestVersion int16
		}
	}
)

// Addr returns the address of the SMTP server with the format host:port
func (smtp *SmtpConfig) Addr() string {
	return fmt.Sprintf("%s:%d", smtp.Host, smtp.Port)
}

// From returns the email address of the sender with the format username@host
func (smtp *SmtpConfig) From() string {
	return fmt.Sprintf("%s@%s", smtp.FromName, smtp.FromUrl)
}

func bindStruct(field *reflect.StructField, tag string) {
	for j := range field.Type.NumField() {
		subField := field.Type.Field(j)

		if env := subField.Tag.Get("mapstructure"); env != "" {
			if subField.Type.Kind() == reflect.Struct {
				bindStruct(&subField, tag+"."+env)
				continue
			}
			env = tag + "." + env
			log.Debug().Str("env", env).Msg("binding env")
			_ = viper.BindEnv(env)
		}
	}
}

func bindAllEnv() {
	t := reflect.TypeOf(Config{}) //nolint:all

	for i := range t.NumField() {
		field := t.Field(i)
		if env := field.Tag.Get("mapstructure"); env != "" && field.Type.Kind() != reflect.Struct {
			log.Debug().Str("env", env).Msg("binding env")
			_ = viper.BindEnv(env)
		}

		if field.Type.Kind() == reflect.Struct {
			bindStruct(&field, field.Tag.Get("mapstructure"))
		}
	}
}

func Load() (Config, error) {
	c := new(Config)

	viper.SetConfigType("env")
	viper.SetEnvPrefix("HUBBLE")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	// Default values
	viper.SetDefault("port", 3288)
	viper.SetDefault("app_url", "http://localhost:3288")
	viper.SetDefault("plugins.directory", ".plugins")
	viper.SetDefault("driver.kv", kv.DriverBadgerDb)
	viper.SetDefault("environment", EnvironmentDevelopment)
	viper.SetDefault("search.mode", SearchModeThreshold.String())
	viper.SetDefault("search.threshold", 30.0)

	bindAllEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, seer.Wrap("read_config_file", err, "failed to read config file")
		}
	}

	if err := viper.Unmarshal(c); err != nil {
		return Config{}, seer.Wrap("unmarshal_config", err, "failed to unmarshal config")
	}

	if err := c.Validate(); err != nil {
		return Config{}, err
	}

	return *c, nil
}

// EnabledEmbeddings if the base URL and the embeddings model are set, then the embeddings are enabled
func (l *LLM) EnabledEmbeddings() bool {
	return l.enabledEmbeddings && l.BaseURL != "" && l.EmbeddingsModel != ""
}

// Debug returns true if the application is running in the development environment or if the debug flag is set
func (c *Config) Debug() bool {
	return (c.Environment == EnvironmentDevelopment || c.Flags.Debug)
}

// InProduction returns true if the application is running in the production environment
func (c *Config) InProduction() bool {
	return (c.Environment == EnvironmentProduction)
}

// InStaging returns true if the application is running in the staging environment
func (c *Config) InStaging() bool {
	return (c.Environment == EnvironmentStaging)
}

// InDevelopment returns true if the application is running in the development environment
func (c *Config) InDevelopment() bool {
	return (c.Environment == EnvironmentDevelopment)
}

// Validate validates the configuration values and ensures they are correct, otherwise returns an error
func (c *Config) Validate() error {
	if c.Port == 0 {
		log.Warn().Msg("port is not set, defaulting to 3288")
		c.Port = 3288
	}

	if c.AppUrl == "" {
		return seer.New("app_url", "app_url is not set, set this as `HUBBLE_APP_URL`")
	}

	if c.Keys.CookieSecret == "" {
		return seer.New(
			"cookie_key",
			"cookie secret key is not set, set this as `HUBBLE_KEY_COOKIE_SECRET`. this is required to sign cookies",
		)
	}

	// Check if the TOTP secrets are set and are in the correct format
	if err := c.validateTotpSecrets(); err != nil {
		return err
	}

	// Ensure the URL is valid and reachable
	if c.LLM.BaseURL != "" {
		parsedURL, err := url.Parse(c.LLM.BaseURL)
		if err != nil {
			return seer.New("llm_base_url", "invalid LLM base URL")
		}

		// Check if the URL is reachable
		if !lib.IsHttpReachable(fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)) {
			log.Warn().Msgf("LLM base URL %s is not reachable", c.LLM.BaseURL)
		}

		// Check if the embeddings model is set
		if c.LLM.EmbeddingsModel != "" {
			c.LLM.enabledEmbeddings = true
		}
	}

	return nil
}

// LatestTotpKeyVersion returns the latest version of the TOTP hash
func (c *Config) LatestTotpKeyVersion() int16 {
	return c.totp.latestVersion
}

// GetLatestTotpKey returns the latest TOTP secret key
func (c *Config) GetLatestTotpKey() string {
	key := c.GetTotpKey(c.LatestTotpKeyVersion())
	if key == nil {
		panic("latest TOTP key is not set")
	}

	return *key
}

// GetTotpKey returns the TOTP secret key for the given version
func (c *Config) GetTotpKey(version int16) *string {
	secret, ok := c.totp.keys[version]
	if !ok {
		return nil
	}

	return &secret
}

func (c *Config) validateTotpSecrets() error {
	if c.totp.keys == nil {
		c.totp.keys = make(map[int16]string)
	}

	if c.Keys.TotpSecrets == "" {
		return seer.New(
			"totp_secrets",
			"TOTP secrets are not set, set this as `HUBBLE_KEY_TOTP_SECRETS`. This is required to sign the TOTP secrets stored in the database. this should be in a similar format to 'v1_xxxx,v2_xxxxx'",
		)
	}

	secrets := strings.Split(c.Keys.TotpSecrets, ",")
	var latestVersion int16

	for _, secret := range secrets {
		if secret == "" {
			return seer.New(
				"totp_secrets",
				fmt.Sprintf("TOTP secret key %s is not in the correct format", secret),
			)
		}

		key, err := parseTotpKey(secret)
		if err != nil {
			return err
		}

		if key.version > latestVersion {
			latestVersion = key.version
		}

		if v, ok := c.totp.keys[key.version]; ok {
			return seer.New(
				"totp_secrets",
				fmt.Sprintf("TOTP secret key %s for version %d is already set", v, key.version),
			)
		}

		c.totp.keys[key.version] = key.secret
	}

	c.totp.latestVersion = latestVersion

	return nil
}

type TotpKey struct {
	version int16
	secret  string
}

func (k TotpKey) String() string {
	return fmt.Sprintf("v%d_%s", k.version, lib.RedactString(k.secret, 4))
}

func (k TotpKey) Version() int   { return int(k.version) }
func (k TotpKey) Secret() string { return k.secret }

func parseTotpKey(secret string) (TotpKey, error) {
	var totpKey TotpKey

	secret = strings.TrimSpace(secret)
	partitionIdx := strings.Index(secret, "_")
	if partitionIdx == -1 {
		return totpKey, seer.New(
			"totp_secret",
			fmt.Sprintf(
				"TOTP secret key %s is not in the correct format, expected format is 'v1_xxxxx'",
				secret,
			),
		)
	}

	version := secret[:partitionIdx]
	key := secret[partitionIdx+1:]

	if version[0] != 'v' {
		return totpKey, seer.New(
			"totp_secret",
			fmt.Sprintf(
				"TOTP secret key %s is not in the correct format, expected format is 'v1_xxxxx'",
				secret,
			),
		)
	}

	if n, err := fmt.Sscanf(version, "v%d", &totpKey.version); err != nil || n != 1 {
		return totpKey, seer.Wrap(
			"totp_secret",
			err,
			"failed to parse TOTP secret version, expected a valid integer",
		)
	}

	if len(key) < 12 {
		return totpKey, seer.New(
			"totp_secret",
			"TOTP secret key is too short, expected at least 12 characters",
		)
	}

	totpKey.secret = key
	return totpKey, nil
}
