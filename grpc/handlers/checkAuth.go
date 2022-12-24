package handlers

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-one/GoStarter/api"
	"github.com/pavel-one/GoStarter/grpc/internal/appauth"
	"github.com/pavel-one/GoStarter/grpc/internal/pgmodels"
)

type CheckAuthService struct {
	api.UnimplementedCheckAuthServiceServer
	BaseDB
}

func NewCheckAuthService(db *sqlx.DB) *CheckAuthService {
	cas := new(CheckAuthService)
	cas.DB = db
	return cas
}

func (c *CheckAuthService) CheckAuth(ctx context.Context, req *api.TokenRequest) (*api.CheckAuthResponse, error) {
	user := new(pgmodels.User)
	token := new(pgmodels.AccessToken)
	claims, err := appauth.GetClaims(req.Token)
	if err != nil {
		if err.(*jwt.ValidationError).Errors == 16 {
			token.GetByToken(c.DB, req.Token)
			if token.ID == 0 {
				return nil, err
			}

			token.Delete(c.DB)
		}
		return nil, err
	}

	if err = user.CheckOnExistsWithoutPassword(c.DB, claims.Unique, claims.Service); err != nil {
		return nil, errors.New("user not found")
	}

	token, err = user.GetTokenByStr(c.DB, req.Token)
	if err != nil {
		return nil, errors.New("token not found")
	}
	if token.ID == 0 {
		return nil, err
	}

	return &api.CheckAuthResponse{UserUUID: user.ID}, nil
}
