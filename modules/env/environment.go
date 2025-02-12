package env

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	Development = "development"
	Staging     = "staging"
	Production  = "production"
)

func LoadEnv() {
	godotenv.Load()
	flavor := getEnvFromArgs()
	godotenv.Overload("." + flavor + ".env")
}

func getEnvFromArgs() string {
	flavorArg := os.Args[1]
	flavor := ""

	if strings.Contains(flavorArg, "flavor=") {
		kv := strings.Split(flavorArg, "=")
		if len(kv) == 2 {
			flavor = kv[1]
		}
	}

	if flavor == Development || flavor == Staging || flavor == Production {
		return flavor
	}

	return Development
}
