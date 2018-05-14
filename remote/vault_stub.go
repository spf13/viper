package remote

// jna -- This stub will connect our vault provider with the
// crypt.ConfigManager interface. The main code is in vault/vault.go
//
// Until I can get the maintainers of the crypt package to support
// vault, we can make vault work this way.

import (
  "io"
  crypt "github.com/xordataexchange/crypt/config"
  vault "github.com/spf13/viper/vault"
)

func NewStandardVaultConfigManager(machines []string) (crypt.ConfigManager, error) {
	store, err := vault.New(machines)
	if err != nil {
		return nil, err
	}
	return crypt.NewStandardConfigManager(store)
}

func NewVaultConfigManager(machines []string, keystore io.Reader) (crypt.ConfigManager, error) {
	store, err := vault.New(machines)
	if err != nil {
		return nil, err
	}
	return crypt.NewConfigManager(store, keystore)
}
