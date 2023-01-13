package client

import (
	"github.com/cherryReptile/WS-AUTH/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceClients struct {
	Conn      *grpc.ClientConn
	App       api.AuthAppServiceClient
	GitHub    api.AuthGithubServiceClient
	Google    api.AuthGoogleServiceClient
	Telegram  api.AuthTelegramServiceClient
	CheckAuth api.CheckAuthServiceClient
	Logout    api.LogoutServiceClient
	Profile   api.ProfileServiceClient
}

func (s *ServiceClients) Init(conn *grpc.ClientConn) {
	s.Conn = conn
	s.App = api.NewAuthAppServiceClient(s.Conn)
	s.GitHub = api.NewAuthGithubServiceClient(s.Conn)
	s.Google = api.NewAuthGoogleServiceClient(s.Conn)
	s.Telegram = api.NewAuthTelegramServiceClient(s.Conn)
	s.CheckAuth = api.NewCheckAuthServiceClient(s.Conn)
	s.Logout = api.NewLogoutServiceClient(s.Conn)
	s.Profile = api.NewProfileServiceClient(s.Conn)
}

func NewConn(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return conn, nil
}