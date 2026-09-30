package meta

import "os"

const (
	envPageID          = "META_PAGE_ID"
	envPageAccessToken = "META_PAGE_ACCESS_TOKEN"
	envAppSecret       = "META_APP_SECRET"
	envVerifyToken     = "META_VERIFY_TOKEN"
)

type Config struct {
	PageID          string
	PageAccessToken string
	AppSecret       string
	VerifyToken     string
}

func NewConfigFromEnv() Config {
	return Config{
		PageID:          os.Getenv(envPageID),
		PageAccessToken: os.Getenv(envPageAccessToken),
		AppSecret:       os.Getenv(envAppSecret),
		VerifyToken:     os.Getenv(envVerifyToken),
	}
}
