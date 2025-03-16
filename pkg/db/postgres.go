package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuanta7/qworker/config"
)

type NotifyMessage struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	ID     uint64 `json:"id"`
}

type PostgresClient struct {
	Pool         *pgxpool.Pool
	QueryBuilder squirrel.StatementBuilderType
}

func NewPostgresClient(cfg *config.Config, opts ...PostgresOption) (*PostgresClient, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.Postgres.GetConnectionString())
	if err != nil {
		return nil, err
	}

	dbConfig.MinConns = 1
	dbConfig.MaxConns = 2

	// overwrite default config
	for _, opt := range opts {
		opt(dbConfig)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{
		Pool:         conn,
		QueryBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func MustNewPostgresClient(cfg *config.Config) *PostgresClient {
	client, err := NewPostgresClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

func (p *PostgresClient) Close() {
	p.Pool.Close()
}

type PostgresOption func(c *pgxpool.Config)

func WithMaxConns(maxConns int32) PostgresOption {
	return func(c *pgxpool.Config) {
		c.MaxConns = maxConns
	}
}

func WithMinConns(minConns int32) PostgresOption {
	return func(c *pgxpool.Config) {
		c.MinConns = minConns
	}
}
