package vault_client

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

type Conf struct {
	SecretToken string
	Address     string
	Mount       string
}

type VaultClient struct {
	cl    *vault.Client
	mount string
}

func New(cfg Conf) (*VaultClient, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating hashicorp vault: %w", err)
	}

	client.SetToken(cfg.SecretToken)

	return &VaultClient{
		cl:    client,
		mount: cfg.Mount,
	}, nil
}

func (sv *VaultClient) ReadSecret(
	ctx context.Context,
	path string,
	dst any,
) error {
	secret, err := sv.cl.KVv2(sv.mount).Get(ctx, path)
	if err != nil {
		return fmt.Errorf("reading secret: %w", err)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           dst,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return fmt.Errorf("creating mapstructure decoder: %w", err)
	}

	err = decoder.Decode(secret.Data)
	if err != nil {
		return fmt.Errorf("decoding secret: %w", err)
	}

	return nil
}
