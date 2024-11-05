package common

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"k8s.io/apiserver/pkg/server/options/encryptionconfig"
)

func readEncryptionConfig(ctx context.Context) (*encryptionconfig.EncryptionConfiguration, error) {
	encryptionConfigPath := os.Getenv("GPTSCRIPT_ENCRYPTION_CONFIG_FILE")
	if encryptionConfigPath == "" {
		var err error
		if encryptionConfigPath, err = xdg.ConfigFile("gptscript/encryptionconfig.yaml"); err != nil {
			return nil, fmt.Errorf("failed to read encryption config from standard location: %w", err)
		}
	}

	if _, err := os.Stat(encryptionConfigPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat encryption config file: %w", err)
	}

	// Use k8s libraries to load the encryption config from the file:
	return encryptionconfig.LoadEncryptionConfig(ctx, encryptionConfigPath, false, "gptscript")
}

func (d Database) encryptCred(ctx context.Context, cred GptscriptCredential) (GptscriptCredential, error) {
	if d.transformer == nil {
		return cred, nil
	}

	secretBytes := []byte(cred.Secret)
	encryptedSecretBytes, err := d.transformer.TransformToStorage(ctx, secretBytes, uid(cred.ServerURL))
	if err != nil {
		return GptscriptCredential{}, fmt.Errorf("failed to encrypt secret: %w", err)
	}
	cred.Secret = fmt.Sprintf("{\"e\": %q}", base64.StdEncoding.EncodeToString(encryptedSecretBytes))

	return cred, nil
}

func (d Database) decryptCred(ctx context.Context, cred GptscriptCredential) (GptscriptCredential, error) {
	if d.transformer == nil {
		return cred, nil
	}

	var secretMap map[string]string
	if err := json.Unmarshal([]byte(cred.Secret), &secretMap); err == nil {
		if encryptedSecretB64, exists := secretMap["e"]; exists && len(secretMap) == 1 {
			encryptedSecretBytes, err := base64.StdEncoding.DecodeString(encryptedSecretB64)
			if err != nil {
				return GptscriptCredential{}, fmt.Errorf("failed to decode secret: %w", err)
			}

			secretBytes, _, err := d.transformer.TransformFromStorage(ctx, encryptedSecretBytes, uid(cred.ServerURL))
			if err != nil {
				return GptscriptCredential{}, fmt.Errorf("failed to decrypt secret: %w", err)
			}
			cred.Secret = string(secretBytes)
		}
	}

	return cred, nil
}
