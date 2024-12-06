package solana

import (
	"context"
	"fmt"

	"mywallet/pkg/logger"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
)

type Client struct {
	client *rpc.Client
	logger *logger.Logger
}

func NewClient(rpcURL string, logger *logger.Logger) *Client {
	return &Client{
		client: rpc.New(rpcURL),
		logger: logger,
	}
}

func (c *Client) GetBalance(ctx context.Context, address string) (decimal.Decimal, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid address: %w", err)
	}

	balance, err := c.client.GetBalance(
		ctx,
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert lamports to SOL (1 SOL = 1e9 lamports)
	solBalance := decimal.NewFromInt(int64(balance.Value)).
		Div(decimal.NewFromInt(1e9))

	return solBalance, nil
}

func (c *Client) Transfer(ctx context.Context, fromPrivateKey solana.PrivateKey, toPublicKey solana.PublicKey, amount decimal.Decimal) (string, error) {
	// Convert SOL to lamports
	lamports := amount.Mul(decimal.NewFromInt(1e9)).IntPart()

	recent, err := c.client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				uint64(lamports),
				fromPrivateKey.PublicKey(),
				toPublicKey,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(fromPrivateKey.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(fromPrivateKey.PublicKey()) {
			return &fromPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	sig, err := c.client.SendTransactionWithOpts(ctx, tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig.String(), nil
}

func (c *Client) GetTransaction(ctx context.Context, signature string) (*rpc.GetTransactionResult, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	tx, err := c.client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}
