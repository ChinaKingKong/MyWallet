package service

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"mywallet/internal/models"
	"mywallet/internal/repository"
	"mywallet/pkg/logger"
	solanaclient "mywallet/pkg/solana"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletService struct {
	logger   *logger.Logger
	solana   *solanaclient.Client
	postgres *repository.PostgresRepository
	redis    *repository.RedisRepository
}

func NewWalletService(
	logger *logger.Logger,
	rpcURL string,
	postgres *repository.PostgresRepository,
	redis *repository.RedisRepository,
) (*WalletService, error) {
	solanaClient := solanaclient.NewClient(rpcURL, logger)

	return &WalletService{
		logger:   logger,
		solana:   solanaClient,
		postgres: postgres,
		redis:    redis,
	}, nil
}

func (s *WalletService) Deposit(ctx context.Context, address string, amount decimal.Decimal) error {
	// 验证金额
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("deposit amount must be greater than 0")
	}

	// 验证地址
	if _, err := solana.PublicKeyFromBase58(address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// 更新数据库余额
	if err := s.postgres.UpdateBalance(ctx, address, amount); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// 创建交易记录
	tx := &models.Transaction{
		ID:          solana.NewWallet().PublicKey().String(), // 使用随机公钥作为ID
		FromWallet:  "deposit",
		ToWallet:    address,
		Amount:      amount,
		Type:        "deposit",
		Status:      "completed",
		CreatedAt:   time.Now(),
		CompletedAt: time.Now(),
	}

	if err := s.postgres.CreateTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 清除缓存
	if err := s.redis.CacheWallet(ctx, nil); err != nil {
		s.logger.Logger.Error("failed to clear wallet cache", zap.Error(err))
	}

	return nil
}

func (s *WalletService) GetBalance(ctx context.Context, address string) (decimal.Decimal, error) {
	// 验证地址
	if _, err := solana.PublicKeyFromBase58(address); err != nil {
		return decimal.Zero, fmt.Errorf("invalid address: %w", err)
	}

	// 先查缓存
	if wallet, err := s.redis.GetCachedWallet(ctx, address); err == nil && wallet != nil {
		return wallet.Balance, nil
	}

	// 查询Solana实时余额
	balance, err := s.solana.GetBalance(ctx, address)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}

	// 更新缓存
	wallet := &models.Wallet{
		Address:   address,
		Balance:   balance,
		UpdatedAt: time.Now(),
	}
	if err := s.redis.CacheWallet(ctx, wallet); err != nil {
		s.logger.Logger.Error("failed to cache wallet", zap.Error(err))
	}

	return balance, nil
}

func (s *WalletService) GetTransactions(ctx context.Context, address string) ([]models.Transaction, error) {
	// 验证地址
	if _, err := solana.PublicKeyFromBase58(address); err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// 先查缓存
	if txs, err := s.redis.GetCachedTransactions(ctx, address); err == nil && txs != nil {
		return txs, nil
	}

	// 查询数据库获取本地交易记录
	dbTxs, err := s.postgres.GetTransactions(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions from database: %w", err)
	}

	// 查询 Solana 链上交易记录
	chainTxs, err := s.solana.GetTransaction(ctx, address)
	if err != nil {
		s.logger.Logger.Error("failed to get transactions from blockchain",
			zap.String("address", address),
			zap.Error(err))
		// 如果链上查询失败，仍返回数据库的记录
		return dbTxs, nil
	}

	// 将链上交易转换为 models.Transaction
	chainTxModels := make([]models.Transaction, 0)
	if chainTxs != nil {
		// 获取交易信息
		if chainTxs.Transaction == nil {
			s.logger.Logger.Warn("transaction is nil")
			return dbTxs, nil
		}

		transaction, err := chainTxs.Transaction.GetTransaction() // 获取交易和错误
		if err != nil {
			s.logger.Logger.Warn("failed to get transaction", zap.Error(err))
			return dbTxs, nil
		}
		message := transaction.Message // 使用获取的交易
		if reflect.DeepEqual(message, solana.Message{}) || len(message.AccountKeys) < 2 {
			s.logger.Logger.Warn("transaction has insufficient accounts")
			return dbTxs, nil
		}

		// 获取交易签名
		signature := transaction.Signatures[0]

		// 获取账户信息
		accounts := message.AccountKeys
		if len(accounts) >= 2 {
			fromAccount := accounts[0].String()
			toAccount := accounts[1].String()

			// 计算转账金额
			amount := decimal.NewFromFloat(float64(chainTxs.Meta.PostBalances[0]-chainTxs.Meta.PreBalances[0]) / 1e9)

			// 获取区块时间
			blockTime := int64(*chainTxs.BlockTime)

			tx := models.Transaction{
				ID:          signature.String(),
				FromWallet:  fromAccount,
				ToWallet:    toAccount,
				Amount:      amount,
				Type:        "transfer",
				Status:      "completed",
				CreatedAt:   time.Unix(blockTime, 0),
				CompletedAt: time.Unix(blockTime, 0),
			}
			chainTxModels = append(chainTxModels, tx)
		}
	}

	// 添加验证函数
	if err := s.validateTransactionData(chainTxs); err != nil {
		s.logger.Logger.Warn("invalid transaction data",
			zap.Error(err))
		return dbTxs, nil
	}

	// 合并链上和数据库的交易记录
	allTxs := make([]models.Transaction, 0, len(dbTxs)+len(chainTxModels))
	allTxs = append(allTxs, dbTxs...)
	allTxs = append(allTxs, chainTxModels...)

	// 更新缓存
	if err := s.redis.CacheTransactions(ctx, address, allTxs); err != nil {
		s.logger.Logger.Error("failed to cache transactions", zap.Error(err))
	}

	return allTxs, nil
}

// 添加验证函数
func (s *WalletService) validateTransactionData(tx *rpc.GetTransactionResult) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	if tx.Transaction == nil {
		return fmt.Errorf("transaction data is nil")
	}

	if tx.Meta == nil {
		return fmt.Errorf("transaction meta is nil")
	}

	if len(tx.Meta.PreBalances) == 0 || len(tx.Meta.PostBalances) == 0 {
		return fmt.Errorf("balance information is missing")
	}

	if tx.BlockTime == nil {
		return fmt.Errorf("block time is missing")
	}

	return nil
}

func (s *WalletService) Transfer(ctx context.Context, fromPrivateKeyStr, toAddress string, amount decimal.Decimal) error {
	// 解析私钥
	fromPrivateKey, err := solana.PrivateKeyFromBase58(fromPrivateKeyStr)
	if err != nil {
		return fmt.Errorf("invalid from private key: %w", err)
	}

	// 获取发送方公钥
	fromAddress := fromPrivateKey.PublicKey().String()

	// 解析接收方地址
	toPubKey, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}

	// 验证转账金额
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be greater than 0")
	}

	// 检查发送方余额
	fromBalance, err := s.GetBalance(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get sender balance: %w", err)
	}
	if fromBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// 调用 Solana 客户端执行实际转账，使用私钥
	signature, err := s.solana.Transfer(ctx, fromPrivateKey, toPubKey, amount)
	if err != nil {
		return fmt.Errorf("failed to execute transfer on blockchain: %w", err)
	}

	// 开始转账操作
	// 1. 扣除发送方余额
	if err := s.postgres.UpdateBalance(ctx, fromAddress, amount.Neg()); err != nil {
		s.logger.Logger.Error("failed to update sender balance",
			zap.String("from", fromAddress),
			zap.Error(err))
		return fmt.Errorf("failed to update sender balance: %w", err)
	}

	// 2. 增加接收方余额
	if err := s.postgres.UpdateBalance(ctx, toAddress, amount); err != nil {
		s.logger.Logger.Error("failed to update receiver balance",
			zap.String("to", toAddress),
			zap.Error(err))
		// 回滚发送方余额
		if rollbackErr := s.postgres.UpdateBalance(ctx, fromAddress, amount); rollbackErr != nil {
			s.logger.Logger.Error("failed to rollback sender balance",
				zap.String("from", fromAddress),
				zap.Error(rollbackErr))
		}
		return fmt.Errorf("failed to update receiver balance: %w", err)
	}

	// 3. 创建交易记录
	tx := &models.Transaction{
		ID:          signature,
		FromWallet:  fromAddress,
		ToWallet:    toAddress,
		Amount:      amount,
		Type:        "transfer",
		Status:      "completed",
		CreatedAt:   time.Now(),
		CompletedAt: time.Now(),
	}

	if err := s.postgres.CreateTransaction(ctx, tx); err != nil {
		s.logger.Logger.Error("failed to create transaction record", zap.Error(err))
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// 4. 清除相关缓存
	if err := s.redis.CacheWallet(ctx, nil); err != nil {
		s.logger.Logger.Error("failed to clear wallet cache", zap.Error(err))
	}

	return nil
}
