package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// API 处理方法
func Deposit(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Amount  string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现存款逻辑
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

func Withdraw(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Amount  string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现提款逻辑
	c.JSON(http.StatusOK, gin.H{"message": "withdrawal successful"})
}

func Transfer(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func GetBalance(c *gin.Context) {
	address := c.Param("address")

	// TODO: 实现余额查询逻辑
	c.JSON(http.StatusOK, gin.H{"address": address, "balance": "0"})
}

func GetTransactions(c *gin.Context) {
	address := c.Param("address")

	// TODO: 实现交易历史查询逻辑
	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"transactions": []string{},
	})
}
