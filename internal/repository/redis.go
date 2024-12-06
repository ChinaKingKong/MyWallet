package repository

import (
	"context"
	"errors"

	"mywallet/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

const (
	balanceKeyPrefix = "wallet:balance:"
	assetsKeyPrefix  = "wallet:assets:"
	addBalanceScript = `
		local balance = redis.call('GET', KEYS[1])
		if not balance then
			redis.call('SET', KEYS[1], ARGV[1])
			return ARGV[1]
		end
		return redis.call('INCRBY', KEYS[1], ARGV[1])
	`
	subBalanceScript = `
		local balance = redis.call('GET', KEYS[1])
		if not balance then
			return nil
		end
		if tonumber(balance) < tonumber(ARGV[1]) then
			return nil
		end
		return redis.call('DECRBY', KEYS[1], ARGV[1])
	`
	transferScript = `
		local fromBalance = redis.call('GET', KEYS[1])
		if not fromBalance then
			return {err = 'no_balance'}
		end
		local fromNum = tonumber(fromBalance)
		local transferNum = tonumber(ARGV[1])
		if fromNum < transferNum then
			return {err = 'insufficient_balance'}
		end
		redis.call('DECRBY', KEYS[1], ARGV[1])
		redis.call('INCRBY', KEYS[2], ARGV[1])
		return 'ok'
	`
)

type RedisRepository struct {
	client *redis.Client
	// 添加 Lua 脚本对象
	addScript      *redis.Script
	subScript      *redis.Script
	transferScript *redis.Script
	logger         *logger.Logger
}

func NewRedisRepository(addr string, logger *logger.Logger) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisRepository{
		client:         client,
		addScript:      redis.NewScript(addBalanceScript),
		subScript:      redis.NewScript(subBalanceScript),
		transferScript: redis.NewScript(transferScript),
		logger:         logger,
	}, nil
}

// Add 增加余额
func (r *RedisRepository) AddBalance(ctx context.Context, address string, amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return errors.New("amount must be positive")
	}

	key := r.getBalanceKey(address)
	amountInt := amount.Mul(decimal.New(1, 8)).IntPart() // 转换为最小单位

	_, err := r.addScript.Run(ctx, r.client, []string{key}, amountInt).Result()
	return err
}

// checkLuaResult 检查 Lua 脚本返回的结果
func (r *RedisRepository) checkLuaResult(result interface{}) error {
	switch result {
	case "ok":
		return nil
	case "no_balance":
		return errors.New("source account has no balance")
	case "insufficient_balance":
		return errors.New("insufficient balance")
	case nil:
		return errors.New("insufficient balance")
	default:
		return errors.New("unknown transfer error")
	}
}

// Sub 减少余额
func (r *RedisRepository) SubBalance(ctx context.Context, address string, amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return errors.New("amount must be positive")
	}

	key := r.getBalanceKey(address)
	amountInt := amount.Mul(decimal.New(1, 8)).IntPart()

	result, err := r.subScript.Run(ctx, r.client, []string{key}, amountInt).Result()
	if err != nil {
		return err
	}
	return r.checkLuaResult(result)
}

// Transfer 在两个账户之间转账
func (r *RedisRepository) Transfer(ctx context.Context, fromAddress, toAddress string, amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return errors.New("amount must be positive")
	}

	if fromAddress == toAddress {
		return errors.New("cannot transfer to self")
	}

	fromKey := r.getBalanceKey(fromAddress)
	toKey := r.getBalanceKey(toAddress)
	amountInt := amount.Mul(decimal.New(1, 8)).IntPart()

	result, err := r.transferScript.Run(ctx, r.client, []string{fromKey, toKey}, amountInt).Result()
	if err != nil {
		return err
	}

	return r.checkLuaResult(result)
}

// GetBalance 获取余额
func (r *RedisRepository) GetBalance(ctx context.Context, address string) (decimal.Decimal, error) {
	key := r.getBalanceKey(address)
	result, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.New(result, -8), nil
}

func (s *RedisRepository) getBalanceKey(address string) string {
	return balanceKeyPrefix + address
}

// 添加获取Redis客户端的方法
func (r *RedisRepository) GetClient() *redis.Client {
	return r.client
}

// SetAssets 设置用户资产缓存
func (r *RedisRepository) SetAssets(ctx context.Context, address string, assets string) error {
	key := r.getAssetsKey(address)
	return r.client.Set(ctx, key, assets, 0).Err()
}

// GetAssets 获取用户资产缓存
func (r *RedisRepository) GetAssets(ctx context.Context, address string) (string, error) {
	key := r.getAssetsKey(address)
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// DeleteAssets 删除用户资产缓存
func (r *RedisRepository) DeleteAssets(ctx context.Context, address string) error {
	key := r.getAssetsKey(address)
	return r.client.Del(ctx, key).Err()
}

// getAssetsKey 获取资产缓存的键
func (r *RedisRepository) getAssetsKey(address string) string {
	return assetsKeyPrefix + address
}
