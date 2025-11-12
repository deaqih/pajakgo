package repository

import (
	"accounting-web/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE username = ? LIMIT 1"
	err := r.db.Get(&user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE email = ? LIMIT 1"
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE id = ? LIMIT 1"
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (name, username, email, password_hash, role, is_active)
	          VALUES (:name, :username, :email, :password_hash, :role, :is_active)`
	result, err := r.db.NamedExec(query, user)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET name = :name, username = :username, email = :email,
	          role = :role, is_active = :is_active WHERE id = :id`
	_, err := r.db.NamedExec(query, user)
	return err
}

func (r *UserRepository) UpdatePassword(id int, passwordHash string) error {
	query := "UPDATE users SET password_hash = ? WHERE id = ?"
	_, err := r.db.Exec(query, passwordHash, id)
	return err
}
