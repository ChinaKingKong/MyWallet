package service

import (
	"context"
	"testing"
	"time"

	"mywallet/internal/models"
	"mywallet/internal/repository"
	"mywallet/pkg/logger"

	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Postgres Repository
type MockPostgresRepository struct {
	mock.Mock
	repository.PostgresRepository
}

func (m *MockPostgresRepository) UpdateBalance(ctx context.Context, address string, amount decimal.Decimal) error {
	args := m.Called(ctx, address, amount)
	return args.Error(0)
}

func (m *MockPostgresRepository) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockPostgresRepository) GetTransactions(ctx context.Context, address string) ([]models.Transaction, error) {
	args := m.Called(ctx, address)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

// Mock Redis Repository
type MockRedisRepository struct {
	mock.Mock
	repository.RedisRepository
}

func (m *MockRedisRepository) CacheWallet(ctx context.Context, wallet *models.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockRedisRepository) GetCachedWallet(ctx context.Context, address string) (*models.Wallet, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockRedisRepository) CacheTransactions(ctx context.Context, address string, txs []models.Transaction) error {
	args := m.Called(ctx, address, txs)
	return args.Error(0)
}

func (m *MockRedisRepository) GetCachedTransactions(ctx context.Context, address string) ([]models.Transaction, error) {
	args := m.Called(ctx, address)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func TestDeposit(t *testing.T) {
	// 设置
	ctx := context.Background()
	logger := logger.NewLogger()
	mockPostgres, err := repository.NewPostgresRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create postgres repository: %v", err) // 错误处理
	}
	mockRedis, err := repository.NewRedisRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create redis repository: %v", err) // 错误处理
	}

	service, err := NewWalletService(logger, "https://api.devnet.solana.com",
		mockPostgres,
		mockRedis)
	assert.NoError(t, err)

	validAddress := solana.NewWallet().PublicKey().String()
	amount := decimal.NewFromFloat(1.0)

	// 执行测试
	err = service.Deposit(ctx, validAddress, amount)
	assert.NoError(t, err)
}

func TestGetBalance(t *testing.T) {
	// 设置
	ctx := context.Background()
	logger := logger.NewLogger()
	mockPostgres, err := repository.NewPostgresRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create postgres repository: %v", err) // 错误处理
	}
	mockRedis, err := repository.NewRedisRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create redis repository: %v", err) // 错误处理
	}

	service, err := NewWalletService(logger, "https://api.devnet.solana.com",
		(*repository.PostgresRepository)(mockPostgres),
		(*repository.RedisRepository)(mockRedis))
	assert.NoError(t, err)

	validAddress := solana.NewWallet().PublicKey().String()
	expectedBalance := decimal.NewFromFloat(1.0)

	// 执行测试
	balance, err := service.GetBalance(ctx, validAddress)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
}

func TestGetTransactions(t *testing.T) {
	// 设置
	ctx := context.Background()
	logger := logger.NewLogger()
	mockPostgres, err := repository.NewPostgresRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create postgres repository: %v", err) // 错误处理
	}
	mockRedis, err := repository.NewRedisRepository("localhost", logger)
	if err != nil {
		t.Fatalf("failed to create redis repository: %v", err) // 错误处理
	}

	service, err := NewWalletService(logger, "https://api.devnet.solana.com",
		(*repository.PostgresRepository)(mockPostgres),
		(*repository.RedisRepository)(mockRedis))
	assert.NoError(t, err)

	validAddress := solana.NewWallet().PublicKey().String()
	expectedTxs := []models.Transaction{
		{
			ID:          "tx1",
			FromWallet:  validAddress,
			ToWallet:    "address2",
			Amount:      decimal.NewFromFloat(1.0),
			Type:        "transfer",
			Status:      "completed",
			CreatedAt:   time.Now(),
			CompletedAt: time.Now(),
		},
	}
	// 执行测试
	txs, err := service.GetTransactions(ctx, validAddress)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxs, txs)
}
