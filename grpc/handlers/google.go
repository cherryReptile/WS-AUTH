package handlers

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-one/GoStarter/api"
	"github.com/pavel-one/GoStarter/internal/appauth"
	"github.com/pavel-one/GoStarter/internal/models"
)

type GoogleAuthService struct {
	api.UnimplementedAuthGoogleServiceServer
	BaseDB
}

func NewGoogleAuthService(db *sqlx.DB) *GoogleAuthService {
	gs := new(GoogleAuthService)
	gs.DB = db
	return gs
}

func (a *GoogleAuthService) Login(ctx context.Context, req *api.GoogleRequest) (*api.AppResponse, error) {
	user := new(models.User)
	token := new(models.AccessToken)

	user.FindByUniqueAndService(a.DB, req.Email, "google")
	if user.ID == 0 {
		user.UniqueRaw = req.Email
		user.AuthorizedBy = "google"
		if err := user.Create(a.DB); err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, errors.New("user not found")
	}

	tokenStr, err := appauth.GenerateToken(user.ID, user.UniqueRaw, user.AuthorizedBy)
	if err != nil {
		return nil, err
	}

	token.Token = tokenStr
	token.UserID = user.ID
	if err = token.Create(a.DB); err != nil {
		return nil, err
	}

	return &api.AppResponse{
		Struct: &api.User{
			ID:           uint64(user.ID),
			UniqueRaw:    user.UniqueRaw,
			AuthorizedBy: user.AuthorizedBy,
			CreatedAt:    user.CreatedAt.String(),
		}, TokenStr: token.Token}, nil
}