package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(acc *Account, isAdmin bool) error
	DeleteAccount(number int64) error
	UpdateAccount(*Account) error
	UpdateAccountBalance(accountNumber int, newBalance int64) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
	CreateTransfer(fromAccount, toAccount, amount int64) error
	GetTransferHistory() ([]*Transfer, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgressStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}
func (s *PostgresStore) createAccountTable() error {
	// query1 := `DROP TABLE IF EXISTS transfer;`
	query := `create table if not exists account(
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		encrypted_password varchar(100),
		balance serial,
		is_admin boolean,
		created_at timestamp
	)`

	transferQuery := `
	CREATE TABLE IF NOT EXISTS transfer (
		id SERIAL PRIMARY KEY,
		from_account INTEGER,
		to_account INTEGER,
		amount INTEGER,
		timestamp TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	// _, err = s.db.Exec(query1)
	// if err != nil {
	// 	return err
	// }
	_, err = s.db.Exec(transferQuery)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account, isAdmin bool) error {

	query := `insert into account 
	(first_name,last_name,number,encrypted_password,balance,is_admin,created_at)
	values ($1,$2,$3,$4,$5,$6,$7)`

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		isAdmin,
		acc.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(account *Account) error {
	query := `
		UPDATE account
		SET is_admin = $1
		WHERE id = $2
	`
	_, err := s.db.Exec(query, account.IsAdmin, account.ID)
	return err
}

func (s *PostgresStore) UpdateAccountBalance(accountNumber int, newBalance int64) error {
	query := `
		UPDATE account
		SET balance = $1
		WHERE number = $2
	`

	_, err := s.db.Exec(query, newBalance, accountNumber)
	return err
}

func (s *PostgresStore) CreateTransfer(fromAccount, toAccount, amount int64) error {
	query := `insert into transfer (from_account, to_account, amount, timestamp)
    values ($1, $2, $3, now())`

	_, err := s.db.Exec(query, fromAccount, toAccount, amount)
	return err
}

func (s *PostgresStore) GetTransferHistory() ([]*Transfer, error) {
	query := `
		SELECT id, from_account, to_account, amount, timestamp
		FROM transfer
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*Transfer{}
	for rows.Next() {
		transaction := new(Transfer)
		err := rows.Scan(&transaction.ID, &transaction.FromAccount, &transaction.ToAccount, &transaction.Amount, &transaction.Timestamp)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *PostgresStore) DeleteAccount(id int64) error {
	_, err := s.db.Query("delete from account where id = $1", id)
	return err
}
func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with number [%d] not found", number)
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *PostgresStore) IsAdminAccount(number int64) (bool, error) {
	query := "SELECT is_admin FROM account WHERE number = $1"

	var isAdmin bool
	err := s.db.QueryRow(query, number).Scan(&isAdmin)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

func (s *PostgresStore) GetAdminAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM account WHERE is_admin = true")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.IsAdmin,
		&account.CreatedAt)

	return account, err
}
