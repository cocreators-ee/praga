package backend

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// CookieAuthConfig contains configuration for cookie authentication
type CookieAuthConfig struct {
	CookieName string `yaml:"cookie_name" validate:"required,min=1,max=64"`
	Domain     string `yaml:"domain" validate:"required,min=1,max=255"`
	Secure     bool   `yaml:"secure" validate:""`
}

// ServerConfig contains configuration for the HTTP server
type ServerConfig struct {
	ListenType string `yaml:"listen_type" validate:"required,oneof=unix http"`
	Socket     string `yaml:"socket" validate:"min=1,max=255"`
	Host       string `yaml:"host" validate:"min=1,max=255"`
	Port       int    `yaml:"port" validate:"gte=1,lte=65535"`
}

// RateLimitConfigItem contains details for rate limiting
type RateLimitConfigItem struct {
	PerHour int `yaml:"per_hour" validate:"gte=0,lte=100000"`
}

// RateLimitConfig defines rate limiting for different things
type RateLimitConfig struct {
	IP    RateLimitConfigItem `yaml:"ip"`
	Email RateLimitConfigItem `yaml:"email"`
}

// JWTConfig configures the authentication token properties
type JWTConfig struct {
	ValidSeconds int `yaml:"valid_seconds" validate:"required,gte=1,lte=1576800000"`
}

// AuthConfig changes how authentication works
type AuthConfig struct {
	Mode      string          `yaml:"mode" validate:"required,oneof=email"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// EmailConfig configures email related settings
type EmailConfig struct {
	ValidDomains  []string `yaml:"valid_domains" validate:"dive,min=1,max=255"`
	ValidEmails   []string `yaml:"valid_emails" validate:"dive,email"`
	EmailProvider string   `yaml:"email_provider" validate:"required,oneof=mailjet"`
	From          string   `yaml:"from" validate:"required,email"`
	FromName      string   `yaml:"from_name" validate:"required,min=1"`
}

// MailjetConfig provides Mailjet API configuration
type MailjetConfig struct {
	APIKeyPublic  string `yaml:"apikey_public" validate:"min=0,max=255"`
	APIKeyPrivate string `yaml:"apikey_private" validate:"min=0,max=255"`
}

// Config provides all the configuration parsed from praga.yaml
type Config struct {
	Title      string           `yaml:"title" validate:"min=1,max=64"`
	Brand      string           `yaml:"brand" validate:"min=1,max=64"`
	Support    string           `yaml:"support" validate:"min=1,max=255"`
	SigningKey string           `yaml:"signing_key" validate:"required,min=16,max=64"`
	CookieAuth CookieAuthConfig `yaml:"cookie_auth"`
	Auth       AuthConfig       `yaml:"auth"`
	Email      EmailConfig      `yaml:"email"`
	Mailjet    MailjetConfig    `yaml:"mailjet"`
	Server     ServerConfig     `yaml:"server"`
	JWT        JWTConfig        `yaml:"jwt"`
}

// LoadConfig loads a praga.yaml file and parses it into a Config
func LoadConfig(configPath string) (bool, Config) {
	c := &Config{}
	c.Title = "Login"
	c.Brand = "Private Area"
	c.Support = "technical support"
	c.CookieAuth.CookieName = "PRAGA_TOKEN"
	c.CookieAuth.Secure = true
	c.Auth.Mode = "email"
	c.Email.EmailProvider = "mailjet"
	c.JWT.ValidSeconds = 86400
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 8086
	c.Server.ListenType = "http"

	f, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(f, c); err != nil {
		log.Fatal(err)
	}

	if c.SigningKey == "openssl rand -hex 32" {
		log.Fatal("Generate a new signing_key in the configuration e.g. with: openssl rand -hex 32")
	}

	// Allow MJ_APIKEY_PRIVATE and MJ_APIKEY_PUBLIC environment overrides
	mkAPIKeyPrivate := os.Getenv("MJ_APIKEY_PRIVATE")
	if mkAPIKeyPrivate != "" {
		c.Mailjet.APIKeyPrivate = mkAPIKeyPrivate
	}

	mjAPIKeyPublic := os.Getenv("MJ_APIKEY_PUBLIC")
	if mjAPIKeyPublic != "" {
		c.Mailjet.APIKeyPublic = mjAPIKeyPublic
	}

	if c.Email.EmailProvider == "mailjet" {
		if c.Mailjet.APIKeyPublic == "" || c.Mailjet.APIKeyPrivate == "" {
			log.Fatal("Mailjet provider missing API key configuration")
		}
	}

	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(c); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, verr := range validationErrors {
				log.Print(strings.Replace(verr.Error(), "Config.", "", 1))
			}
		} else {
			log.Print(err)
		}
		return false, *c
	}

	return true, *c
}
