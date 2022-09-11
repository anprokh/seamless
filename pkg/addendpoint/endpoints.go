package addendpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"seamless/pkg/addservice"
)

type Endpoints struct {
	GetBalanceEndpoint          endpoint.Endpoint
	WithdrawAndDepositEndpoint  endpoint.Endpoint
	RollbackTransactionEndpoint endpoint.Endpoint
}

func MakeEndpoints(svc addservice.SeamlessService) Endpoints {
	return Endpoints{
		GetBalanceEndpoint:          makeGetBalanceEndpoint(svc),
		WithdrawAndDepositEndpoint:  makeWithdrawAndDepositEndpoint(svc),
		RollbackTransactionEndpoint: makeRollbackTransactionEndpoint(svc),
	}
}

func makeGetBalanceEndpoint(svc addservice.SeamlessService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(addservice.GetBalanceRequest)
		res, err := svc.GetBalance(ctx, req)
		if err != nil {
			res.ErrorCode = 32001
		}
		return res, nil
	}
}

func makeWithdrawAndDepositEndpoint(svc addservice.SeamlessService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(addservice.WithdrawAndDepositRequest)
		res, err := svc.WithdrawAndDeposit(ctx, req)
		if err != nil {
			res.ErrorCode = 32001
		}
		return res, nil
	}
}

func makeRollbackTransactionEndpoint(svc addservice.SeamlessService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(addservice.RollbackTransactionRequest)
		res, err := svc.RollbackTransaction(ctx, req)
		if err != nil {
			res.ErrorCode = 32001
		}
		return res, nil
	}
}
