package store

import (
	"database/sql"
	"time"

	_ "golang.org/x/crypto/bcrypt"
)

type UserStore interface{
	CreateUser(*User) error
	GetUserByUsername(username string) (*User, error)
	UpdateUser(*User) error
}

type password struct {
	plainText *string
	hash []byte
}

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash password `json:"-"` // `json:"-"` means to ignore the value in the struct
	Bio          string `json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

func (pg *PostgresUserStore) CreateUser(user *User) error {
	query := `
		INSERT INTO users
			(
				username, 
				email, 
				password_hash, 
				bio
			)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at;
	`

	err := pg.db.QueryRow(
		query, 
		user.Username, 
		user.PasswordHash.hash, 
		user.Bio,
	).Scan(
		&user.ID, 
		&user.CreatedAt, 
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgresUserStore) GetUserByUsername(username string) (*User, error) {
	user :=  &User{
		PasswordHash: password{},
	}

	query := `
	SELECT
		id,
		username,
		email,
		password_hash,
		bio,
		created_at,
		updated
	FROM users 
	WHERE username = $1;`
	err := pg.db.QueryRow(
		query, 
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresUserStore) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET
			username = $1,
			email = $2,
			bio = $3,
			updated_at CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at;
	`

	result, err := pg.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Bio,
		user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}