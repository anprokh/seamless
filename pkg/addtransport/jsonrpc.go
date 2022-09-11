package addtransport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/transport/http/jsonrpc"
	endpoints "seamless/pkg/addendpoint"
	"seamless/pkg/addservice"
)

func NewJSONRPCHandler(endpoints endpoints.Endpoints) *jsonrpc.Server {
	handler := jsonrpc.NewServer(
		makeEndpointCodecMap(endpoints),
	)
	return handler
}

func makeEndpointCodecMap(endpoints endpoints.Endpoints) jsonrpc.EndpointCodecMap {
	return jsonrpc.EndpointCodecMap{
		"getBalance": jsonrpc.EndpointCodec{
			Endpoint: endpoints.GetBalanceEndpoint,
			Decode:   decodeGetBalanceRequest,
			Encode:   encodeGetBalanceResponse,
		},
		"withdrawAndDeposit": jsonrpc.EndpointCodec{
			Endpoint: endpoints.WithdrawAndDepositEndpoint,
			Decode:   decodeWithdrawAndDepositRequest,
			Encode:   encodeWithdrawAndDepositResponse,
		},
		"rollbackTransaction": jsonrpc.EndpointCodec{
			Endpoint: endpoints.RollbackTransactionEndpoint,
			Decode:   decodeRollbackTransactionRequest,
			Encode:   encodeRollbackTransactionResponse,
		},
	}
}

func decodeGetBalanceRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req addservice.GetBalanceRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to GetBalance request: %s", err),
		}
	}
	return req, nil
}

func encodeGetBalanceResponse(_ context.Context, obj interface{}) (json.RawMessage, error) {
	res, ok := obj.(addservice.GetBalanceResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *GetBalanceResponse failed. Got %T, %+v", obj, obj),
		}
	}
	// handle possible seamless error
	if res.ErrorCode > 0 {
		return nil, seamlessErrorByCode(res.ErrorCode)
	}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal GetBalanceResponse response: %s", err)
	}
	return b, nil
}

func decodeWithdrawAndDepositRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req addservice.WithdrawAndDepositRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to WithdrawAndDeposit request: %s", err),
		}
	}
	return req, nil
}

func encodeWithdrawAndDepositResponse(_ context.Context, obj interface{}) (json.RawMessage, error) {
	res, ok := obj.(addservice.WithdrawAndDepositResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *WithdrawAndDepositResponse failed. Got %T, %+v", obj, obj),
		}
	}
	// handle possible seamless error
	if res.ErrorCode > 0 {
		return nil, seamlessErrorByCode(res.ErrorCode)
	}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal WithdrawAndDeposit response: %s", err)
	}
	return b, nil
}

func decodeRollbackTransactionRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req addservice.RollbackTransactionRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to RollbackTransaction request: %s", err),
		}
	}
	return req, nil
}

func encodeRollbackTransactionResponse(_ context.Context, obj interface{}) (json.RawMessage, error) {
	res, ok := obj.(addservice.RollbackTransactionResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *RollbackTransactionResponse failed. Got %T, %+v", obj, obj),
		}
	}
	// handle possible seamless error
	if res.ErrorCode > 0 {
		return nil, seamlessErrorByCode(res.ErrorCode)
	}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal RollbackTransactionResponse response: %s", err)
	}
	return b, nil
}

// decode errors from business-logic
func seamlessErrorByCode(errorCode int) jsonrpc.Error {
	switch errorCode {
	case 1:
		return jsonrpc.Error{Code: 1, Message: "ErrNotEnoughMoneyCode"}
	case 2:
		return jsonrpc.Error{Code: 2, Message: "ErrIllegalCurrencyCode"}
	case 3:
		return jsonrpc.Error{Code: 3, Message: "ErrNegativeDepositCode"}
	case 4:
		return jsonrpc.Error{Code: 4, Message: "ErrNegativeWithdrawalCode"}
	case 5:
		return jsonrpc.Error{Code: 5, Message: "ErrSpendingBudgetExceeded"}
	case 10:
		return jsonrpc.Error{Code: 10, Message: "ErrCurrencyCodeIsDifferentFromBalance"} // валюта запроса отличается от валюты баланса
	case 11:
		return jsonrpc.Error{Code: 11, Message: "ErrTransactionAlreadyCanceled"} // транзакция была откачена
	case 32001:
		return jsonrpc.Error{Code: 32001, Message: "ErrInternalServerError"} // в тестовом задании не будем детализировать внутренние ошибки
	default:
		return jsonrpc.Error{Code: errorCode, Message: "unknown error"}
	}
}
