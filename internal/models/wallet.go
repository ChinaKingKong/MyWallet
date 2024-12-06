package models

import (
    "time"
    "github.com/shopspring/decimal"
)

type Wallet struct {
    ID        string          `json:"id"`
    Address   string          `json:"address"`
    Balance   decimal.Decimal `json:"balance"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}

type Transaction struct {
    ID          string          `json:"id"`
    FromWallet  string          `json:"from_wallet"`
    ToWallet    string          `json:"to_wallet"`
    Amount      decimal.Decimal `json:"amount"`
    Type        string          `json:"type"` // deposit, withdraw, transfer
    Status      string          `json:"status"`
    CreatedAt   time.Time       `json:"created_at"`
    CompletedAt time.Time       `json:"completed_at"`
} 