package store

import (
	"database/sql"
	"time"

	"github.com/lesi97/internal/tokens"
)

type TokenStore interface{
	Insert(token *tokens.Token) error
	CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokenForUser(userID int, scope string) error
}

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: db,
	}
}

func (pg PostgresTokenStore) Insert(token *tokens.Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)
	`
	_, err := pg.db.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err
}

func (pg PostgresTokenStore) CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = pg.Insert(token)
	return token, err
}

func (pg PostgresTokenStore) DeleteAllTokenForUser(userID int, scope string) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1
		AND user_id = $2
	`
	_, err := pg.db.Exec(query, scope, userID)
	return err
}