package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func envKeys(req CreateAuthRequest) []string {
	keys := []string{}
	if req.IsJwt() {
		keys = append(keys, "JWT_SECRET")
	}

	if !req.IsPassword() {
		keys = append(keys, "OTP_SECRET")
	}

	// TotpConfig.Key is hex decoded and must be exactly 32 bytes
	if req.Mfa {
		keys = append(keys, "TOTP_KEY")
	}

	return keys
}

// .env files are gitignored, so they get real secrets; .env.example is committed and gets blanks.
func appendEnv(req CreateAuthRequest) error {
	keys := envKeys(req)
	if len(keys) == 0 {
		return nil
	}

	secrets := []string{}
	blanks := []string{"# auth secrets: openssl rand -hex 32"}
	for _, key := range keys {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return err
		}

		secrets = append(secrets, key+`="`+hex.EncodeToString(secret)+`"`)
		blanks = append(blanks, key+`=""`)
	}

	if err := appendLines(filepath.Join(req.RootDirPath, ENV), secrets); err != nil {
		return err
	}

	localEnv := filepath.Join(req.RootDirPath, DEPLOYMENTS, LOCAL, ENV)
	if _, err := os.Stat(localEnv); err == nil {
		if err := appendLines(localEnv, secrets); err != nil {
			return err
		}
	}

	return appendLines(filepath.Join(req.RootDirPath, ENV_EXAMPLE), blanks)
}

func appendLines(path string, lines []string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	prefix := ""
	if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n")) {
		prefix = "\n"
	}

	_, err = f.WriteString(prefix + strings.Join(lines, "\n") + "\n")
	return err
}
