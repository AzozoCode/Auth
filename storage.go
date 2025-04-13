package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {

	connStr := "user=postgres dbname=postgres password=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err

	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `Create Table if not exists account(
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance float,
		created_at timestamp default now(),
		updated_at timestamp
	)`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {

	query := `INSERT INTO ACCOUNT (first_name,last_name,number,balance,created_at,updated_at)
	 VALUES($1,$2,$3,$4,$5,$6)`

	resp, err := s.db.Exec(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
		acc.UpdatedAt)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {

	query := `SELECT * FROM ACCOUNT`

	resp, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for resp.Next() {

		account, err := scanIntoAccount(resp)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)

	}

	return accounts, nil

}
func (s *PostgresStore) DeleteAccount(id int) error {
	query := `DELETE FROM ACCOUNT WHERE id = $1`
	_, err := s.db.Exec(query, id)

	if err != nil {
		return fmt.Errorf("an error occurred deleting record with id:%d\nerror:%+v", id, err)
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {

	query := `SELECT * FROM ACCOUNT WHERE ID = $1`
	result, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}

	for result.Next() {

		return scanIntoAccount(result)

	}

	return nil, fmt.Errorf("account %d not found", id)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {

	account := new(Account)

	err := rows.Scan(&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt)

	return account, err
}
