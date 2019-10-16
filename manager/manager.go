package manager

import (
	"database/sql"
	"errors"
	"fmt"

	"gopkg.in/validator.v2"

	"github.com/natacha0923/user_balance_service/structs"
)

var (
	ErrInsufficientBalance = errors.New("amount is bigger than balance")
	ErrChanged             = errors.New("balance was not changed")
)

type UserBalanceManager struct {
	Db *sql.DB
}

func (m *UserBalanceManager) ChangeBalance(rq structs.ChangeBalanceRequest) error {
	err := validator.Validate(rq.Body)
	if err != nil {
		return err
	}

	return m.changeBalance(rq.Body.UserId, rq.Body.Amount)
}

func (m *UserBalanceManager) changeBalance(userId uint64, delta int64) error {
	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	// get user and lock
	row := m.Db.QueryRow("SELECT user_id, balance FROM user_balance WHERE user_id = $1 FOR UPDATE", userId)
	var userBalance structs.UserBalance
	err = row.Scan(&userBalance.UserId, &userBalance.Balance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return fmt.Errorf("can not found user with id = %d", userId)
		}
		return err
	}

	if userBalance.Balance+delta < 0 {
		tx.Rollback()
		return ErrInsufficientBalance
	}

	// update balance
	result, err := m.Db.Exec("UPDATE user_balance SET balance = balance + $1 where user_id = $2", delta, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		tx.Rollback()
		return ErrChanged
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (m *UserBalanceManager) Transfer(rq structs.TransferRequest) error {
	err := validator.Validate(rq.Body)
	if err != nil {
		return err
	}

	return m.transfer(rq.Body.From, rq.Body.To, rq.Body.Amount)
}

func (m *UserBalanceManager) transfer(fromId, toId uint64, amount int64) error {
	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	// get users and lock
	rows, err := m.Db.Query("SELECT user_id, balance FROM user_balance WHERE user_id = $1 OR user_id = $2 FOR UPDATE", fromId, toId)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var from, to structs.UserBalance
	for rows.Next() {
		var bk structs.UserBalance
		err := rows.Scan(&bk.UserId, &bk.Balance)
		if err != nil {
			tx.Rollback()
			return err
		}

		switch bk.UserId {
		case fromId:
			from = bk
		case toId:
			to = bk
		}

	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return err
	}

	if from.IsEmpty() {
		tx.Rollback()
		return fmt.Errorf("can not found user with id = %d", fromId)
	}

	if to.IsEmpty() {
		tx.Rollback()
		return fmt.Errorf("can not found user with id = %d", toId)
	}

	if from.Balance-amount < 0 {
		tx.Rollback()
		return ErrInsufficientBalance
	}

	// Transferring from
	result, err := m.Db.Exec("UPDATE user_balance SET balance = balance - $1 where user_id = $2", amount, fromId)
	if err != nil {
		tx.Rollback()
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		tx.Rollback()
		return ErrChanged
	}

	// Transferring to
	result, err = m.Db.Exec("UPDATE user_balance SET balance = balance + $1 where user_id = $2", amount, toId)
	if err != nil {
		tx.Rollback()
		return err
	}

	count, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		tx.Rollback()
		return ErrChanged
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
