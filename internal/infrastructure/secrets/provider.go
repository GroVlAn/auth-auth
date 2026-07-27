package secrets

import (
	"context"
	"fmt"
)

type vaultClient interface {
	ReadSecret(
		ctx context.Context,
		path string,
		dst any,
	) error
}

type Paths struct {
	Token  string
	Redis  string
	Hasher string
}

type SecretsProvider struct {
	client vaultClient
	paths  Paths
}

func New(client vaultClient, paths Paths) *SecretsProvider {
	return &SecretsProvider{
		client: client,
		paths:  paths,
	}
}

func (sp *SecretsProvider) Load(ctx context.Context) (*Secrets, error) {
	var secrets Secrets

	secretLoaders := []struct {
		name string
		path string
		dst  any
	}{
		{
			name: "tokens",
			path: sp.paths.Token,
			dst:  &secrets.Token,
		},
		{
			name: "redis",
			path: sp.paths.Redis,
			dst:  &secrets.Redis,
		},
		{
			name: "hasher",
			path: sp.paths.Hasher,
			dst:  &secrets.Hasher,
		},
	}

	for _, loader := range secretLoaders {
		if err := sp.client.ReadSecret(
			ctx,
			loader.path,
			loader.dst,
		); err != nil {
			return nil, fmt.Errorf(
				"reading %s secrets: %w",
				loader.name,
				err,
			)
		}
	}

	return &secrets, nil
}
