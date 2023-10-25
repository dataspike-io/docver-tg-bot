package cache

import (
	"context"
	"errors"
	"github.com/Yiling-J/theine-go"
	dataspike "github.com/dataspike-io/docver-sdk-go"
)

const size = 1000

type MemoryCache struct {
	client *theine.Cache[string, *dataspike.Verification]
}

func (m *MemoryCache) GetVerification(ctx context.Context, tgId string) (*dataspike.Verification, error) {
	if value, ok := m.client.Get(tgId); !ok {
		return nil, errors.New("verification not found")
	} else {
		return value, nil
	}
}

func (m *MemoryCache) SetVerification(ctx context.Context, tgId string, v *dataspike.Verification) error {
	if !m.client.Set(tgId, v, 0) {
		return errors.New("set error")
	}

	return nil
}

func (m *MemoryCache) RemoveVerification(ctx context.Context, tgId string) error {
	m.client.Delete(tgId)
	return nil
}

func NewMemoryCache(maxSize int64) (*MemoryCache, error) {
	if maxSize <= 0 {
		maxSize = size
	}
	client, err := theine.NewBuilder[string, *dataspike.Verification](maxSize).Build()
	if err != nil {
		return nil, err
	}
	return &MemoryCache{client: client}, nil
}
