// Copyright 2018-2020 Celer Network

package webapi

import (
	"context"
	"errors"

	"github.com/celer-network/goCeler/celersdk"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/webapi/rpc"
	"google.golang.org/grpc"
)

type InternalApiServer struct {
	*ApiServer
}

func NewInternalApiServer(
	webPort int,
	grpcPort int,
	allowedOrigins string,
	keystore string,
	password string,
	dataPath string,
	config string,
	extSigner bool) *InternalApiServer {
	apiServer := NewApiServer(webPort, grpcPort, allowedOrigins, keystore, password, dataPath, config, extSigner)
	return &InternalApiServer{apiServer}
}

func (s *InternalApiServer) Start() {
	gs := grpc.NewServer()
	rpc.RegisterWebApiServer(gs, s.ApiServer)
	rpc.RegisterInternalWebApiServer(gs, s)
	s.ApiServer.serve(gs)
}

func (s *InternalApiServer) OpenTrustedPaymentChannel(
	context context.Context, request *rpc.OpenPaymentChannelRequest) (*rpc.ChannelID, error) {
	callbackImpl := s.callbackImpl
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		go s.apiClient.TcbOpenETHChannel(
			request.PeerAmount,
			s.callbackImpl)
	case entity.TokenType_ERC20:
		go s.apiClient.TcbOpenTokenChannel(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			request.PeerAmount,
			s.callbackImpl)
	default:
		return nil, errors.New("Unknown token type")
	}
	select {
	case cid := <-callbackImpl.channelOpened:
		return &rpc.ChannelID{ChannelId: cid}, nil
	case errMsg := <-callbackImpl.openChannelError:
		return nil, errors.New(errMsg)
	}
}

func (s *InternalApiServer) InstantiateTrustedPaymentChannel(
	context context.Context, request *rpc.TokenInfo) (*rpc.ChannelID, error) {
	var ercType string
	if request.TokenType == entity.TokenType_ETH {
		ercType = ""
	} else {
		ercType = "ERC20"
	}
	callbackImpl := s.callbackImpl
	go s.apiClient.InstantiateChannelForToken(
		&celersdk.Token{
			Erctype: ercType,
			Addr:    request.TokenAddress,
		},
		callbackImpl)
	select {
	case cid := <-callbackImpl.channelOpened:
		return &rpc.ChannelID{ChannelId: cid}, nil
	case errMsg := <-callbackImpl.openChannelError:
		return nil, errors.New(errMsg)
	}
}

func (s *InternalApiServer) DepositNonBlocking(
	context context.Context, request *rpc.DepositOrWithdrawRequest) (*rpc.DepositOrWithdrawJob, error) {
	cb := &depositCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}
	var err error
	var jobID string
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		jobID, err = s.apiClient.DepositETH(request.Amount, cb)
	case entity.TokenType_ERC20:
		jobID, err = s.apiClient.DepositERC20(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			request.Amount,
			cb)
	default:
		err = errors.New("Unknown token type")
	}
	if err != nil {
		return nil, err
	}
	return &rpc.DepositOrWithdrawJob{JobId: jobID}, nil
}

func (s *InternalApiServer) CooperativeWithdrawNonBlocking(
	context context.Context,
	request *rpc.DepositOrWithdrawRequest) (*rpc.DepositOrWithdrawJob, error) {
	cb := &withdrawCallback{
		apiClient: s.apiClient,
		txHash:    make(chan string),
		err:       make(chan string),
	}

	var err error
	var jobID string
	tokenInfo := request.TokenInfo
	switch entity.TokenType(tokenInfo.TokenType) {
	case entity.TokenType_ETH:
		jobID, err = s.apiClient.WithdrawETH(request.Amount, cb)
	case entity.TokenType_ERC20:
		jobID, err = s.apiClient.WithdrawERC20(
			&celersdk.Token{Erctype: "ERC20", Addr: tokenInfo.TokenAddress},
			request.Amount,
			cb)
	default:
		err = errors.New("Unknown token type")
	}
	if err != nil {
		return nil, err
	}
	return &rpc.DepositOrWithdrawJob{JobId: jobID}, nil
}
