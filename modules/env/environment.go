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

var Flavor string = Development

func LoadEnv() {
	godotenv.Load()
	flavor := getEnvFromArgs()
	godotenv.Overload("." + flavor + ".env")
}

func getEnvFromArgs() string {
	flavorArg := os.Args[1]

	if strings.Contains(flavorArg, "flavor=") {
		kv := strings.Split(flavorArg, "=")
		if len(kv) == 2 {
			Flavor = kv[1]
		}
	}

	if Flavor == Development || Flavor == Staging || Flavor == Production {
		return Flavor
	}

	Flavor = Development
	return Flavor
}
