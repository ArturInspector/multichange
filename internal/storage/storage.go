package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dechat/exchange-service/internal/models"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateDeposit(deposit *models.Deposit) error
	GetDepositByAddress(chain models.Chain, address string) (*models.Deposit, error)
	UpdateDeposit(deposit *models.Deposit) error
	CreateWithdrawal(withdrawal *models.Withdrawal) error
	GetPendingWithdrawals(chain models.Chain, limit int) ([]*models.Withdrawal, error)
	UpdateWithdrawal(withdrawal *models.Withdrawal) error
	GetHotWallet(chain models.Chain) (*models.HotWallet, error)
	UpdateHotWalletBalance(chain models.Chain, balance string) error
	Close() error
}

type PostgresStorage struct {
	db *sql.DB
}

func New(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// Deposit methods
func (s *PostgresStorage) CreateDeposit(deposit *models.Deposit) error {
	query := `
		INSERT INTO deposits (chain, address, user_id, order_id, expected_amount, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return s.db.QueryRow(
		query,
		deposit.Chain,
		deposit.Address,
		deposit.UserID,
		deposit.OrderID,
		deposit.ExpectedAmount,
		deposit.Status,
	).Scan(&deposit.ID, &deposit.CreatedAt)
}

func (s *PostgresStorage) GetDepositByAddress(chain models.Chain, address string) (*models.Deposit, error) {
	deposit := &models.Deposit{}
	query := `
		SELECT id, chain, address, user_id, order_id, expected_amount, received_amount,
		       tx_hash, block_number, confirmations, status, created_at, confirmed_at
		FROM deposits
		WHERE chain = $1 AND address = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := s.db.QueryRow(query, chain, address).Scan(
		&deposit.ID,
		&deposit.Chain,
		&deposit.Address,
		&deposit.UserID,
		&deposit.OrderID,
		&deposit.ExpectedAmount,
		&deposit.ReceivedAmount,
		&deposit.TxHash,
		&deposit.BlockNumber,
		&deposit.Confirmations,
		&deposit.Status,
		&deposit.CreatedAt,
		&deposit.ConfirmedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return deposit, err
}

func (s *PostgresStorage) UpdateDeposit(deposit *models.Deposit) error {
	query := `
		UPDATE deposits
		SET received_amount = $1, tx_hash = $2, block_number = $3,
		    confirmations = $4, status = $5, confirmed_at = $6
		WHERE id = $7
	`
	_, err := s.db.Exec(
		query,
		deposit.ReceivedAmount,
		deposit.TxHash,
		deposit.BlockNumber,
		deposit.Confirmations,
		deposit.Status,
		deposit.ConfirmedAt,
		deposit.ID,
	)
	return err
}

// Withdrawal methods
func (s *PostgresStorage) CreateWithdrawal(withdrawal *models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (chain, order_id, from_address, to_address, amount, fee, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return s.db.QueryRow(
		query,
		withdrawal.Chain,
		withdrawal.OrderID,
		withdrawal.FromAddress,
		withdrawal.ToAddress,
		withdrawal.Amount,
		withdrawal.Fee,
		withdrawal.Status,
	).Scan(&withdrawal.ID, &withdrawal.CreatedAt)
}

func (s *PostgresStorage) GetPendingWithdrawals(chain models.Chain, limit int) ([]*models.Withdrawal, error) {
	query := `
		SELECT id, chain, order_id, from_address, to_address, amount, fee,
		       tx_hash, status, block_number, confirmations, created_at, sent_at, confirmed_at
		FROM withdrawals
		WHERE chain = $1 AND status = 'pending'
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := s.db.Query(query, chain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*models.Withdrawal
	for rows.Next() {
		w := &models.Withdrawal{}
		err := rows.Scan(
			&w.ID,
			&w.Chain,
			&w.OrderID,
			&w.FromAddress,
			&w.ToAddress,
			&w.Amount,
			&w.Fee,
			&w.TxHash,
			&w.Status,
			&w.BlockNumber,
			&w.Confirmations,
			&w.CreatedAt,
			&w.SentAt,
			&w.ConfirmedAt,
		)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, rows.Err()
}

func (s *PostgresStorage) UpdateWithdrawal(withdrawal *models.Withdrawal) error {
	query := `
		UPDATE withdrawals
		SET tx_hash = $1, status = $2, block_number = $3,
		    confirmations = $4, sent_at = $5, confirmed_at = $6
		WHERE id = $7
	`
	_, err := s.db.Exec(
		query,
		withdrawal.TxHash,
		withdrawal.Status,
		withdrawal.BlockNumber,
		withdrawal.Confirmations,
		withdrawal.SentAt,
		withdrawal.ConfirmedAt,
		withdrawal.ID,
	)
	return err
}

// HotWallet methods
func (s *PostgresStorage) GetHotWallet(chain models.Chain) (*models.HotWallet, error) {
	wallet := &models.HotWallet{}
	query := `
		SELECT id, chain, address, encrypted_key, balance, last_checked_at
		FROM hot_wallets
		WHERE chain = $1
	`
	err := s.db.QueryRow(query, chain).Scan(
		&wallet.ID,
		&wallet.Chain,
		&wallet.Address,
		&wallet.EncryptedKey,
		&wallet.Balance,
		&wallet.LastCheckedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return wallet, err
}

func (s *PostgresStorage) UpdateHotWalletBalance(chain models.Chain, balance string) error {
	query := `
		UPDATE hot_wallets
		SET balance = $1, last_checked_at = $2
		WHERE chain = $3
	`
	_, err := s.db.Exec(query, balance, time.Now(), chain)
	return err
}
