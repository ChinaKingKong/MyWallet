package repository

import (
	"context"
	"encoding/json"
	"time"

	"mywallet/internal/models"
	"mywallet/pkg/logger"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
	logger *logger.Logger
}

func NewRedisRepository(addr string, logger *logger.Logger) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisRepository{
		client: client,
		logger: logger,
	}, nil
}

func (r *RedisRepository) CacheWallet(ctx context.Context, wallet *models.Wallet) error {
	data, err := json.Marshal(wallet)
	if err != nil {
		return err
	}

	key := "wallet:" + wallet.Address
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *RedisRepository) GetCachedWallet(ctx context.Context, address string) (*models.Wallet, error) {
	key := "wallet:" + address
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var wallet models.Wallet
	err = json.Unmarshal(data, &wallet)
	return &wallet, err
}

func (r *RedisRepository) CacheTransactions(ctx context.Context, address string, txs []models.Transaction) error {
	data, err := json.Marshal(txs)
	if err != nil {
		return err
	}

	key := "transactions:" + address
	return r.client.Set(ctx, key, data, 1*time.Hour).Err()
}

func (r *RedisRepository) GetCachedTransactions(ctx context.Context, address string) ([]models.Transaction, error) {
	key := "transactions:" + address
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction
	err = json.Unmarshal(data, &transactions)
	return transactions, err
}
