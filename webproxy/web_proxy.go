// Copyright 2018-2019 Celer Network

package webproxy

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/rs/cors"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/utils"
	proxyrpc "github.com/celer-network/goCeler-oss/webproxy/rpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type WebProxy struct {
	port                         int
	serverNetworkAddress         string
	mu                           sync.Mutex
	sessionToClientConnectionMap map[string]*clientConnection
}

type clientConnection struct {
	sessionToken string
	rpcClient    rpc.RpcClient
	stream       rpc.Rpc_CelerStreamClient
}

func NewProxy(port int, serverNetworkAddress string) *WebProxy {
	return &WebProxy{
		port:                         port,
		serverNetworkAddress:         serverNetworkAddress,
		sessionToClientConnectionMap: make(map[string]*clientConnection),
	}
}

func (p *WebProxy) Start() {
	gs := grpc.NewServer()
	proxyrpc.RegisterWebProxyRpcServer(gs, p)

	errChan := make(chan error)

	wrappedSvr := grpcweb.WrapServer(gs)
	addr := ":" + strconv.Itoa(p.port)
	httpSvr := &http.Server{
		Addr: addr,
		Handler: cors.New(cors.Options{
			AllowedHeaders:   []string{"*"},
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		}).Handler(wrappedSvr),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       86400 * time.Second,
	}
	log.Infoln("Serving Celer Web Proxy on", addr)
	go func() {
		errChan <- httpSvr.ListenAndServe()
	}()
	err := <-errChan
	gs.GracefulStop()
	log.Fatal(err)
}

func (p *WebProxy) CreateSession(
	ctx context.Context, _ *empty.Empty) (*proxyrpc.SessionToken, error) {
	conn, err := grpc.Dial(
		p.serverNetworkAddress,
		utils.GetClientTlsOption(),
		grpc.WithBlock(),
		grpc.WithTimeout(4*time.Second),
		grpc.WithKeepaliveParams(config.KeepAliveClientParams))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	sessionToken := uuid.New().String()
	p.mu.Lock()
	p.sessionToClientConnectionMap[sessionToken] = &clientConnection{
		sessionToken: sessionToken,
		rpcClient:    rpc.NewRpcClient(conn),
	}
	p.mu.Unlock()
	return &proxyrpc.SessionToken{Token: sessionToken}, nil
}

func (p *WebProxy) OpenChannel(
	ctx context.Context, request *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	c, err := p.getClientConnection(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	response, err := c.rpcClient.CelerOpenChannel(
		context.Background(),
		request)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return response, nil
}

func (p *WebProxy) SubscribeMessages(
	authReq *rpc.AuthReq, proxyStream proxyrpc.WebProxyRpc_SubscribeMessagesServer) error {
	c, err := p.getClientConnection(proxyStream.Context())
	if err != nil {
		log.Error(err)
		return err
	}
	celerStream, err := c.rpcClient.CelerStream(context.Background())
	if err != nil {
		log.Error(err)
		return err
	}
	sendErr := celerStream.Send(&rpc.CelerMsg{
		Message: &rpc.CelerMsg_AuthReq{AuthReq: authReq},
	})
	if sendErr != nil {
		log.Error(sendErr)
		return sendErr
	}
	c.stream = celerStream
	for {
		message, err := celerStream.Recv()
		if err != nil {
			log.Error(err)
			return err
		}
		proxyStream.Send(message)
	}
}

func (p *WebProxy) SendMessage(ctx context.Context, message *rpc.CelerMsg) (*empty.Empty, error) {
	c, err := p.getClientConnection(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if c.stream == nil {
		missingStreamErr := errors.New("Missing stream")
		log.Error(missingStreamErr)
		return nil, missingStreamErr
	}
	err = c.stream.Send(message)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (p *WebProxy) getClientConnection(ctx context.Context) (*clientConnection, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	session := md["session"]
	if len(session) == 0 {
		missingTokenErr := errors.New("Missing session token")
		log.Error(missingTokenErr)
		return nil, missingTokenErr
	}
	sessionToken := md["session"][0]
	p.mu.Lock()
	clientConnection, ok := p.sessionToClientConnectionMap[sessionToken]
	p.mu.Unlock()
	if !ok {
		unknownSessionErr := errors.New("Unknown session")
		return nil, unknownSessionErr
	}
	return clientConnection, nil
}
