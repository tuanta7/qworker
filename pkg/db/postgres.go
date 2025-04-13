package db

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuanta7/qworker/config"
	"log"
)

type PostgresClient interface {
	Pool() *pgxpool.Pool
	QueryBuilder() squirrel.StatementBuilderType
	Close()
}

type postgresClient struct {
	PgxPool    *pgxpool.Pool
	SQLBuilder squirrel.StatementBuilderType
}

func (p *postgresClient) Pool() *pgxpool.Pool {
	return p.PgxPool
}

func (p *postgresClient) QueryBuilder() squirrel.StatementBuilderType {
	return p.SQLBuilder
}

func (p *postgresClient) Close() {
	p.PgxPool.Close()
}

func NewPostgresClient(cfg *config.Config, opts ...PostgresOption) (PostgresClient, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.Postgres.GetConnectionString())
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(dbConfig)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	return &postgresClient{
		PgxPool:    conn,
		SQLBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func MustNewPostgresClient(cfg *config.Config, opts ...PostgresOption) PostgresClient {
	client, err := NewPostgresClient(cfg, opts...)
	if err != nil {
		log.Fatalf("NewPostgresClient(): %s", err)
	}
	return client
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
