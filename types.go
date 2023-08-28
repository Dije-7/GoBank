package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type LoginRequest struct {
	Number   int64 `json:"number"`
	Password string
}

type TransferRequest struct {
	FromAccount int64 `json:"fromAccount"`
	ToAccount   int64 `json:"toAccount"`
	Amount      int64 `json:"amount"`
}

type Transfer struct {
	ID          int64
	FromAccount int64
	ToAccount   int64
	Amount      int64
	Timestamp   time.Time
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
	Balance   int64  `json:"balance"`
	IsAdmin   bool   `json:"isAdmin"`
}
type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           int64     `json:"balance"`
	IsAdmin           bool      `json:"isAdmin"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (a *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(pw)) == nil
}

func NewAccount(firstName, lastName, password string, balance int64) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName:         firstName,
		LastName:          lastName,
		EncryptedPassword: string(encpw),
		Number:            int64(rand.Intn(1000000)),
		Balance:           balance,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
