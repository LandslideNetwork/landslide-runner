package internal

import (
	"context"
	"errors"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// executeQuery performs the query and unmarshals the response
func (s *ChainService) executeQuery(queryPath string, req interface{}, res interface{}) error {
	queryArgs, err := encoding.GetCodec(proto.Name).Marshal(req)
	if err != nil {
		s.log.Fatal("error marshaling request: %v", zap.Error(err))
		return err
	}

	resABCIQuery, err := s.c.ABCIQuery(context.Background(), queryPath, queryArgs)
	if err != nil {
		s.log.Fatal("ABCIQuery failed", zap.Error(err))
		return err
	}

	if resABCIQuery.Response.IsErr() {
		s.log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return errors.New(resABCIQuery.Response.Log)
	}

	if resABCIQuery.Response.Code != 0 {
		s.log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return errors.New(resABCIQuery.Response.Log)
	}

	err = encoding.GetCodec(proto.Name).Unmarshal(resABCIQuery.Response.Value, res)
	if err != nil {
		s.log.Error("Failed to unmarshal response: %v", zap.Error(err))
		return err
	}

	return nil
}

// GetBalances queries the balances of an address
func (s *ChainService) GetBalances(address string) {
	var (
		queryPath = "/cosmos.bank.v1beta1.Query/AllBalances"
		req       = &bank.QueryAllBalancesRequest{
			Address: address,
		}
		res = &bank.QueryAllBalancesResponse{}
	)

	err := s.executeQuery(queryPath, req, res)
	if err != nil {
		return
	}

	s.log.Info("Balance query success",
		zap.String("address", address),
		zap.String("balances", res.Balances.String()),
	)
}

// GetAccountInfo queries the account info of an address
func (s *ChainService) GetAccountInfo(address string) (uint64, error) {
	var (
		queryPath = "/cosmos.auth.v1beta1.Query/AccountInfo"
		req       = &authv1beta1.QueryAccountInfoRequest{
			Address: address,
		}
		res = &authv1beta1.QueryAccountInfoResponse{}
	)

	err := s.executeQuery(queryPath, req, res)
	if err != nil {
		return 0, err
	}

	return res.Info.Sequence, nil
}

// UpdateAccountSequence updates the account sequence of an address
func (s *ChainService) UpdateAccountSequence(signerName string) error {
	acc, ok := s.client.GetAccount(signerName)
	if !ok {
		s.log.Fatal("account not found", zap.String("signerName", signerName))
		return errors.New("account not found")
	}

	seq, err := s.GetAccountInfo(acc.Address)
	if err != nil {
		return err
	}

	if err := s.client.SetSequence(signerName, seq); err != nil {
		return err
	}

	return nil
}
