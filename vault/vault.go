package vault

/* Vault implements Hashicorp-vault based storage for configurations
 * which is substaintally more secure than storing configs in
 * consul or flat files.
 *
 * If using approle authentication. set your environment variables
 * as follows to use this backend
 *
 * export VAULT_SECRET_ID= ... secret ...
 * export VAULT_ROLE_ID= ... role id ...
 * -- or --
 * export VAULT_TOKEN = ....
 *
 * If you are using SSL with vault, you can set
 * export VAULT_CACERT= ... pem file containing ca cert ...
 *           and/or
 * export VAULT_SSL_VERIFY=no
 */

import (
	"fmt"
	"os"
	"time"

	"github.com/xordataexchange/crypt/backend"

	vaultapi "github.com/hashicorp/vault/api"
)

type Client struct {
	client         *vaultapi.Client
	secret         string        // used only with role authentication, nil if using env-VAULT_TOKEN
	secret_ttl     time.Duration // if non-zero, it expires at this time
	secret_acq_at  float64       // when we got the secret
	secret_expires bool
}

func (c *Client) acquireToken(role string, secret string) (string, error) {
	secretData := map[string]interface{}{
		"role_id":   role,
		"secret_id": secret,
	}

	data, err := c.client.Logical().Write("auth/approle/login", secretData)
	if data == nil {
		return "", err
	}
	/* data is now of type *api.Secret and we can use it to set the client up */
	token, err := data.TokenID()
	if err == nil {
		c.client.SetToken(token)
	}

	/* handle expiry */
	ttl, err := data.TokenTTL()
	if err == nil {
		c.secret_ttl = ttl
		if ttl != 0 {
			c.secret_expires = true
		}
	}

	c.secret_acq_at = float64(time.Now().Unix())

	fmt.Println("Got token %s with expiry %d and acquired at %v", token, c.secret_ttl, c.secret_acq_at)
	return token, err
}

// this can be called before operations to ensure token is current
func (c *Client) renewToken() (string, error) {
	if c.secret_expires {
		if (c.secret_ttl.Seconds()+c.secret_acq_at > float64(time.Now().Unix())) && c.secret_ttl != 0 {
			return c.acquireToken(os.Getenv("VAULT_ROLE_ID"), os.Getenv("VAULT_SECRET_ID"))
		} else {
			return "", nil
		}
	} else {
		return "", nil
	}
}

func New(machines []string) (*Client, error) {
	/* default config reads from the environment and sets defaults */
	/* a call to vaultapi.ReadEnvironment is not necessary here. */
	/*
	 * vault environment variables are required to proceed.
	 * either VAULT_TOKEN or VAULT_ROLE_ID and VAULT_SECRET_ID must be set
	 * see: https://github.com/hashicorp/vault/blob/master/api/client.go
	 */

	conf := vaultapi.DefaultConfig()

	if len(machines) > 0 {
		conf.Address = machines[0]
	}

	// from the vault docs -
	// https://godoc.org/github.com/hashicorp/vault/api#Secret
	// If the environment variable `VAULT_TOKEN` is present, the token
	// will be automatically added to the client. Otherwise, you must
	// manually call `SetToken()`.
	var returnval *Client

	client, err := vaultapi.NewClient(conf)

	if err != nil {
		return nil, err
	}

	/* what token are we using? */
	if v := os.Getenv(vaultapi.EnvVaultToken); v == "" {
		/* not using VAULT_TOKEN! */
		if v := os.Getenv("VAULT_ROLE_ID"); v == "" {
			fmt.Fprintf(os.Stderr, "neither VAULT_TOKEN or a VAULT_ROLE_ID/VAULT_SECRET_ID are set. Can't auth to vault.\n")
			return nil, fmt.Errorf("Can't Auth to Vault")
		}
		if v := os.Getenv("VAULT_SECRET_ID"); v == "" {
			fmt.Fprintf(os.Stderr, "VAULT_ROLE_ID set but VAULT_SECRET_ID is empty. Can't auth to vault.\n")
			return nil, fmt.Errorf("Can't Auth to Vault")
		}

		returnval = &Client{client, "", 0, float64(time.Now().Unix()), false}

		/* using the approle secrets, try to acquire a token */
		_, err := returnval.acquireToken(os.Getenv("VAULT_ROLE_ID"), os.Getenv("VAULT_SECRET_ID"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Vault ROLE/SECRET authentication failed - %v\n", err)
			return nil, fmt.Errorf("Can't Auth to Vault")
		}
	} else {
		/* we'll just go ahead with VAULT_TOKEN for auth */
		returnval = &Client{client, os.Getenv(vaultapi.EnvVaultToken), 0, float64(time.Now().Unix()), false}
	}

	return returnval, nil
}

func (c *Client) Get(key string) ([]byte, error) {
	/* note that the vault client only connects when Get is issued if
	 * you are using VAULT_TOKEN authentication (set in the environment)
	 *
	 * If using role authentication, we'll try to acquire a token at init.
	 *
	 * This interface returns only one value from a secret. It expects that the
	 * referenced secret will have the data in the "value" key.
	 */
	data, err := c.client.Logical().Read(key)
	if err != nil {
		fmt.Println("Error during Vault Get -", err)
		return []byte{}, err
	}
	if data.Data == nil {
		return []byte{}, fmt.Errorf("Key ( %s ) was not found.", key)
	}

	v := data.Data["value"].(string)
	return []byte(v), nil
}

func (c *Client) List(key string) (backend.KVPairs, error) {
	// TODO: NOT IMPLEMENTED
	//pairs, err := c.client.Logical().List(key)
	return nil, nil
}

func (c *Client) Set(key string, value []byte) error {
	secretData := map[string]interface{}{
		"value": value,
	}
	_, err := c.client.Logical().Write(key, secretData)

	return err
}

func (c *Client) Watch(key string, stop chan bool) <-chan *backend.Response {
	respChan := make(chan *backend.Response, 0)
	go func() {
		for {
			data, err := c.client.Logical().Read(key)
			if data == nil && err == nil {
				err = fmt.Errorf("Key ( %s ) was not found.", key)
			}
			if err != nil {
				respChan <- &backend.Response{nil, err}
				time.Sleep(time.Second * 5)
				continue
			}

			respChan <- &backend.Response{data.Data["value"].([]byte), nil}
		}
	}()
	return respChan
}
