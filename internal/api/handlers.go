package api

import (
	"net/http"

	"mywallet/internal/config"
	"mywallet/internal/repository"
	"mywallet/internal/service"
	"mywallet/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Server struct {
	cfg    *config.Config
	logger *logger.Logger
	wallet *service.WalletService
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	postgres, err := repository.NewPostgresRepository(cfg.PostgresURL, logger)
	if err != nil {
		logger.Fatal("failed to create postgres repository", err)
		return nil
	}
	redis, err := repository.NewRedisRepository(cfg.RedisURL, logger)
	if err != nil {
		wallet, s_err := service.NewWalletService(logger, cfg.SolanaRPC, postgres, redis)
		if s_err != nil {
			s := &Server{
				cfg:    cfg,
				logger: logger,
				wallet: wallet,
			}
			return s
		} else {
			logger.Fatal("failed to create wallet service", err)
			return nil
		}

	} else {
		logger.Fatal("failed to create redis repository", err)
		return nil
	}
}

// API 处理方法
func (s *Server) Deposit(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Amount  string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现存款逻辑
	amount, err := decimal.NewFromString(req.Amount) // 将字符串转换为 decimal.Decimal
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}
	if err := s.wallet.Deposit(c.Request.Context(), req.Address, amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

func (s *Server) Withdraw(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Amount  string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现提款逻辑
	amount, err := decimal.NewFromString(req.Amount) // 将字符串转换为 decimal.Decimal
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}
	if err := s.wallet.Withdraw(c.Request.Context(), req.Address, amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "withdrawal successful"})
}

func (s *Server) Transfer(c *gin.Context) {
	var req struct {
		FromAddress string `json:"from_address" binding:"required"`
		ToAddress   string `json:"to_address" binding:"required"`
		Amount      string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现转账逻辑
	amount, err := decimal.NewFromString(req.Amount) // 将字符串转换为 decimal.Decimal
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}
	if err := s.wallet.Transfer(c.Request.Context(), req.FromAddress, req.ToAddress, amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func (s *Server) GetBalance(c *gin.Context) {
	address := c.Param("address")

	// TODO: 实现余额查询逻辑
	balance, err := s.wallet.GetBalance(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"address": address, "balance": balance})
}

func (s *Server) GetTransactions(c *gin.Context) {
	address := c.Param("address")

	// TODO: 实现交易历史查询逻辑
	transactions, err := s.wallet.GetTransactions(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"transactions": transactions,
	})
}
