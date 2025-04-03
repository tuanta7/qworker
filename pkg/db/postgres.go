package db

import (
	"context"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuanta7/qworker/config"
)

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

func MustNewPostgresClient(cfg *config.Config, opts ...PostgresOption) *PostgresClient {
	client, err := NewPostgresClient(cfg, opts...)
	if err != nil {
		log.Fatalf("NewPostgresClient(): %s", err)
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
