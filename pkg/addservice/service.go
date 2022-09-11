package addservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
	"time"
)

type SeamlessService interface {
	GetBalance(ctx context.Context, r GetBalanceRequest) (GetBalanceResponse, error)
	WithdrawAndDeposit(ctx context.Context, r WithdrawAndDepositRequest) (WithdrawAndDepositResponse, error)
	RollbackTransaction(ctx context.Context, r RollbackTransactionRequest) (RollbackTransactionResponse, error)
}

type seamlessService struct {
	DB              *sql.DB
	mtxRef          sync.RWMutex
	transactionRefs map[string]bool
}

func (s seamlessService) StoreRef(transactionRef string) error {
	s.mtxRef.Lock()
	defer s.mtxRef.Unlock()
	s.transactionRefs[transactionRef] = true
	return nil
}

func (s seamlessService) FindRef(transactionRef string) bool {
	s.mtxRef.RLock()
	defer s.mtxRef.RUnlock()
	if _, ok := s.transactionRefs[transactionRef]; ok {
		return true
	}
	return false
}

func (s seamlessService) FreeRef(transactionRef string) error {
	s.mtxRef.Lock()
	defer s.mtxRef.Unlock()
	delete(s.transactionRefs, transactionRef)
	return nil
}

func NewSeamlessService(connStr string) (seamlessService, error) {
	var s seamlessService
	time.Sleep(1 * time.Second)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return s, errors.New(fmt.Sprintf("SeamlessService creation failure: '%s'", err))
	}
	err = db.Ping()
	if err != nil {
		return s, errors.New(fmt.Sprintf("SeamlessService creation failure: '%s'", err))
	}
	s.DB = db
	s.transactionRefs = make(map[string]bool)
	return s, nil
}

func (s seamlessService) GetBalance(_ context.Context, r GetBalanceRequest) (GetBalanceResponse, error) {
	var res GetBalanceResponse

	// в рамках тестового задания не будем проверять код валюты, ограничимся проверкой его наличия
	if len(r.Currency) != 3 {
		res.ErrorCode = 2
		return res, nil
	}

	balance, currency, err := s.GetBalanceDB(r.PlayerName)
	if err != nil {
		return res, err
	}
	// обработаем кейс когда валюта запроса баланса отличается от валюты баланса игрока
	if (len(currency) > 0) && (currency != r.Currency) {
		res.ErrorCode = 10
		return res, nil
	}
	res.Balance = balance
	return res, nil
}

func (s seamlessService) WithdrawAndDeposit(_ context.Context, r WithdrawAndDepositRequest) (WithdrawAndDepositResponse, error) {
	var res WithdrawAndDepositResponse

	// ждем завершения предыдущего дублирующего запроса при его наличии
	for {
		if !s.FindRef(r.TransactionRef) {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	_ = s.StoreRef(r.TransactionRef)
	defer s.FreeRef(r.TransactionRef)

	// сразу проверим deposit/withdraw чтобы не делать лишних запросов к БД
	if r.Deposit < 0 {
		res.ErrorCode = 3
		return res, nil
	}
	if r.Withdraw < 0 {
		res.ErrorCode = 4
		return res, nil
	}

	// в рамках тестового задания не будем проверять код валюты, ограничимся проверкой его наличия
	if len(r.Currency) != 3 {
		res.ErrorCode = 2
		return res, nil
	}

	// проверим существование transactionRef (дублирующий запрос)
	var id int
	var txExists, completed, canceled bool

	rows, err := s.DB.Query("select id, completed, canceled from transactions where transactionref = $1 limit 1", r.TransactionRef)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &completed, &canceled)
		if err != nil {
			return res, err
		}
		txExists = true
	}

	if txExists {
		if canceled {
			res.ErrorCode = 11 // транзакция была откачена ранее
			return res, nil
		}
		if completed { // если был успешный дублирующий запрос просто вернем текущий баланс
			balance, _, err := s.GetBalanceDB(r.PlayerName)
			if err != nil {
				return res, err
			}
			res.NewBalance = balance
			res.TransactionId = r.TransactionRef
			return res, nil
		}
		// если транзакция не была ни откачена, ни успешно выполнена - пробуем выполнить повторно
	} else {
		// transactionRef не существует, зарегистрируем транзакцию
		_, err := s.DB.Exec("insert into transactions (playername, withdraw, deposit, currency, transactionref) values ($1, $2, $3, $4, $5)",
			r.PlayerName, r.Withdraw, r.Deposit, r.Currency, r.TransactionRef)
		if err != nil {
			return res, err
		}
	}

	// баланс на начало транзакции (и его валюта)
	balance, currency, err := s.GetBalanceDB(r.PlayerName)
	if err != nil {
		return res, err
	}

	// обработаем кейс когда валюта withdrawAndDeposit отличается от валюты баланса игрока
	if (len(currency) > 0) && (currency != r.Currency) {
		res.ErrorCode = 10
		return res, nil
	}

	// проверим баланс при списании
	newBalance := balance - r.Withdraw
	if newBalance < 0 {
		res.ErrorCode = 1
		return res, nil
	}
	newBalance = newBalance + r.Deposit

	err = s.SetBalanceDB(r.PlayerName, r.Currency, newBalance)
	if err != nil {
		return res, err
	}

	_, err = s.DB.Exec("update transactions set completed = true where transactionref = $1", r.TransactionRef)
	if err != nil {
		return res, err
	}

	res.NewBalance = newBalance
	res.TransactionId = r.TransactionRef // возвращаем входящий т.к. учет своих не ведем
	res.ErrorCode = 0

	return res, nil
}

func (s seamlessService) RollbackTransaction(_ context.Context, r RollbackTransactionRequest) (RollbackTransactionResponse, error) {
	var res RollbackTransactionResponse

	// ждем завершения предыдущего запроса с таким же transactionRef при его наличии
	for {
		if !s.FindRef(r.TransactionRef) {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	_ = s.StoreRef(r.TransactionRef)
	defer s.FreeRef(r.TransactionRef)

	// проверим существование transactionRef
	var id, withdraw, deposit int
	var playername string
	var txExists, completed, canceled bool

	rows, err := s.DB.Query("select id, playername, withdraw, deposit, completed, canceled from transactions where transactionref = $1 limit 1", r.TransactionRef)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &playername, &withdraw, &deposit, &completed, &canceled)
		if err != nil {
			return res, err
		}
		txExists = true
	}

	if !txExists {
		// если транзакция не существует - регистрируем ее как откаченную
		_, err := s.DB.Exec("insert into transactions (playername, transactionref, canceled) values ($1, $2, true)", r.PlayerName, r.TransactionRef)
		if err != nil {
			return res, err
		}
		return res, nil
	} else {
		if canceled { // транзакция была откачена ранее
			return res, nil
		}
		if !completed { // транзакция не была завершена
			return res, nil
		}
	}
	// транзакция существует, завершена, и не была откачена ранее - откатываем

	// текущий баланс
	balance, currency, err := s.GetBalanceDB(playername)
	if err != nil {
		return res, err
	}

	// проверим баланс при списании
	newBalance := balance - deposit
	if newBalance < 0 {
		res.ErrorCode = 1
		return res, nil
	}
	newBalance = newBalance + withdraw

	err = s.SetBalanceDB(playername, currency, newBalance)
	if err != nil {
		return res, err
	}

	_, err = s.DB.Exec("update transactions set canceled = true where transactionref = $1", r.TransactionRef)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s seamlessService) GetBalanceDB(playerName string) (int, string, error) {
	var bal int
	var cur string
	rows, err := s.DB.Query("select balance, currency from balances where playername = $1 limit 1", playerName)
	if err != nil {
		return bal, cur, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&bal, &cur)
		if err != nil {
			return bal, cur, err
		}
	}
	return bal, cur, nil
}

func (s seamlessService) SetBalanceDB(playerName string, currency string, balance int) error {
	_, err := s.DB.Exec("insert into balances (playername, currency, balance) values ($1, $2, $3) on conflict (playername) do update set balance = $3", playerName, currency, balance)
	if err != nil {
		return err
	}
	return nil
}

// В рамках тестового задания ограничимся обязательными полями
type WithdrawAndDepositRequest struct {
	CallerId       int    `json:"callerId"`
	PlayerName     string `json:"playerName"`
	Withdraw       int    `json:"withdraw"`
	Deposit        int    `json:"deposit"`
	Currency       string `json:"currency"`
	TransactionRef string `json:"transactionRef"`
}

type WithdrawAndDepositResponse struct {
	NewBalance    int    `json:"newBalance"`
	TransactionId string `json:"transactionId"`
	ErrorCode     int    `json:"-"`
}

type GetBalanceRequest struct {
	CallerId   int    `json:"callerId"`
	PlayerName string `json:"playerName"`
	Currency   string `json:"currency"`
}

type GetBalanceResponse struct {
	Balance   int `json:"balance"`
	ErrorCode int `json:"-"`
}

type RollbackTransactionRequest struct {
	CallerId       int    `json:"callerId"`
	PlayerName     string `json:"playerName"`
	TransactionRef string `json:"transactionRef"`
}

type RollbackTransactionResponse struct {
	ErrorCode int `json:"-"`
}
