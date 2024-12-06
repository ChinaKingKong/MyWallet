package repository

import (
	"context"
	"database/sql"
	"time"

	"mywallet/internal/models"
	"mywallet/pkg/logger"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type PostgresRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewPostgresRepository(connStr string, logger *logger.Logger) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *PostgresRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	query := `
        INSERT INTO wallets (id, address, balance, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID,
		wallet.Address,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateBalance(ctx context.Context, address string, amount decimal.Decimal) error {
	query := `
        UPDATE wallets 
        SET balance = balance + $1, updated_at = $2
        WHERE address = $3
    `

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), address)
	return err
}

func (r *PostgresRepository) GetBalance(ctx context.Context, address string) (decimal.Decimal, error) {
	query := `SELECT balance FROM wallets WHERE address = $1`

	var balance decimal.Decimal
	err := r.db.QueryRowContext(ctx, query, address).Scan(&balance)
	if err == sql.ErrNoRows {
		return decimal.Zero, nil
	}
	return balance, err
}

func (r *PostgresRepository) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	query := `
        INSERT INTO transactions (id, from_wallet, to_wallet, amount, type, status, created_at, completed_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.ExecContext(ctx, query,
		tx.ID,
		tx.FromWallet,
		tx.ToWallet,
		tx.Amount,
		tx.Type,
		tx.Status,
		tx.CreatedAt,
		tx.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) GetTransactions(ctx context.Context, address string) ([]models.Transaction, error) {
	query := `
        SELECT id, from_wallet, to_wallet, amount, type, status, created_at, completed_at
        FROM transactions
        WHERE from_wallet = $1 OR to_wallet = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.FromWallet,
			&tx.ToWallet,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
			&tx.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}
