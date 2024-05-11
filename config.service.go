package common

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	PROD    = "prod"
	DEV     = "dev"
	STAGING = "staging"
)

var ENV = PROD

func getEnvArgument() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return PROD
}

func isAlreadyLoaded() bool {
	env := os.Getenv("ENV")
	isLoaded := env == PROD || env == DEV || env == STAGING
	if isLoaded {
		setEnv(env)
		GetLogger().Info("Environment already loaded: " + env)
	}
	return isLoaded
}

func LoadEnv() {
	if isAlreadyLoaded() {
		return
	}
	env := getEnvArgument()
	setEnv(env)
	envFileName := "." + ENV + ".env"
	GetLogger().Info("Loading environment: " + envFileName)
	err := godotenv.Load(envFileName)
	if err != nil {
		GetLogger().Fatal(
			"failed to load environment",
			zap.Error(err),
			zap.String("env", env),
			zap.Stack("stack trace"),
		)
	}
}

func setEnv(env string) {
	if env == DEV {
		ENV = DEV
	} else if env == STAGING {
		ENV = STAGING
	} else {
		ENV = PROD
	}
}

func IsDev() bool {
	return ENV == DEV
}

func IsProd() bool {
	return ENV == PROD
}

func IsStaging() bool {
	return ENV == STAGING
}

func ConfigEssentials() {
	LoadEnv()
}

func CleanUp() {
	if logger != nil {
		logger.Sync()
	}
}
