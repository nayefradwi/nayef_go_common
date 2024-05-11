package common

import (
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func GenerateUuid() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		GetLogger().Error("error generating uuid", zap.Error(err))
		return "", err
	}
	uuidStr := strings.ReplaceAll(u.String(), "-", "")
	return uuidStr, nil
}
