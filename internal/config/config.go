package config

type Config struct {
    PostgresURL string
    RedisURL    string
    SolanaRPC   string
    ServerPort  string
}

func Load() (*Config, error) {
    return &Config{
        PostgresURL: "postgres://user:password@localhost:5432/wallet?sslmode=disable",
        RedisURL:    "redis://localhost:6379",
        SolanaRPC:   "https://api.mainnet-beta.solana.com",
        ServerPort:  ":8080",
    }, nil
} 