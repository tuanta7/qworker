package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuanta7/qworker/config"
)

type PostgresClient struct {
	Pool    *pgxpool.Pool
	Builder squirrel.StatementBuilderType
}

func NewPostgresClient(cfg *config.Config) (*PostgresClient, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.Postgres.GetConnectionString())
	if err != nil {
		return nil, err
	}

	dbConfig.MinConns = 3
	dbConfig.MaxConns = 10

	conn, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{
		Pool:    conn,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (p *PostgresClient) Close() {
	p.Pool.Close()
}
