package snapconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/snapcore/snapd/client"
	"github.com/snapcore/snapd/dirs"
)

var (
	ErrSnapctl       = errors.New("snapctl system call failed")
	ErrJSONUnmarshal = errors.New("couldn't unmarhsal json")
)

// This variable copied from
// Vogt, Michael, et al. “Snapd.” Github.com, 2.54.4, Snapcore, 25 Nov. 2020,
// https://github.com/snapcore/snapd/blob/67901c9dd5c22f4ece95b53f43fb0acc73e81a23/cmd/snapctl/main.go.
// Accessed 8 Mar. 2022
var clientConfig = client.Config{
	// snapctl should not try to read $HOME/.snap/auth.json, this will
	// result in apparmor denials and configure task failures
	// (LP: #1660941)
	DisableAuth: true,

	// we need the less privileged snap socket in snapctl
	Socket: dirs.SnapSocket,
}

// This function modified from
// Vogt, Michael, et al. “Snapd.” Github.com, 2.54.4, Snapcore, 25 Nov. 2020,
// https://github.com/snapcore/snapd/blob/67901c9dd5c22f4ece95b53f43fb0acc73e81a23/cmd/snapctl/main.go.
// Accessed 8 Mar. 2022.
func runSnapCtl(args []string) (stdout []byte, stderr []byte, err error) {
	cli := client.New(&clientConfig)
	cookie := os.Getenv("SNAP_COOKIE")
	// for compatibility, if re-exec is not enabled and facing older snapd.
	if cookie == "" {
		cookie = os.Getenv("SNAP_CONTEXT")
	}

	return cli.RunSnapctl(&client.SnapCtlOptions{
		ContextID: cookie,
		Args:      args,
	}, nil)
}

func get(key string) (string, error) {
	stdout, _, err := runSnapCtl([]string{"get", "-d", key})
	if err != nil {
		return "", fmt.Errorf("%w for %s: %v", ErrSnapctl, key, err)
	}
	var result map[string]string
	if err = json.Unmarshal(stdout, &result); err != nil {
		return "", fmt.Errorf("%w for %s: %v", ErrJSONUnmarshal, key, err)
	}
	if len(result) == 0 {
		return "", nil
	}
	return result[key], nil
}

type snapConfigOverrider struct {
	ErrHandler func(error)
}

func (o *snapConfigOverrider) Get(lowerCaseKey string) (string, bool) {
	result, err := get(lowerCaseKey)
	if err != nil {
		o.ErrHandler(err)
	}
	return result, err == nil && len(result) > 0
}

//nolint:revive // we dont want to give access to snapConfigOverrider type without constructor
func Overrider(errHandler func(error)) *snapConfigOverrider {
	return &snapConfigOverrider{
		ErrHandler: errHandler,
	}
}
