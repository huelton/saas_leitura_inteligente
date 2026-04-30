package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID    string
	Name  string
	Email string
	Plan  string
}

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, plan) VALUES ($1,$2,$3,$4,'free')`,
		id, name, email, string(hash))
	if err != nil {
		return nil, err
	}
	return &User{ID: id, Name: name, Email: email, Plan: "free"}, nil
}

func (r *UserRepository) Authenticate(ctx context.Context, email, password string) (*User, error) {
	var u User
	var hash string
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, email, password_hash, plan FROM users WHERE email=$1`, email).Scan(
		&u.ID, &u.Name, &u.Email, &hash, &u.Plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("credenciais inválidas")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errors.New("credenciais inválidas")
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, email, plan FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("usuário não encontrado")
	}
	return &u, err
}
