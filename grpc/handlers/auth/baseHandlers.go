package auth

import (
	"context"
	"errors"
	"github.com/cherryReptile/WS-AUTH/api"
	"github.com/cherryReptile/WS-AUTH/domain"
	"github.com/cherryReptile/WS-AUTH/internal/authtoken"
	"github.com/cherryReptile/WS-AUTH/internal/github"
	"github.com/cherryReptile/WS-AUTH/internal/google"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type BaseHandler struct {
	userUsecase           domain.UserUsecase
	providerUsecase       domain.ProviderUsecase
	tokenUsecase          domain.AuthTokenUsecase
	providersDataUsecase  domain.ProvidersDataUsecase
	usersProvidersUsecase domain.UsersProvidersUsecase
}

type BaseOAuthHandler struct {
	DB     *sqlx.DB
	Config *oauth2.Config
	BaseHandler
	Provider string
}

func (h *BaseOAuthHandler) GetTokenDefault(req *api.OAuthCodeRequest) (*api.OAuthTokenResponse, error) {
	tok, err := h.Config.Exchange(context.Background(), req.Code)
	if err != nil {
		return nil, err
	}
	return &api.OAuthTokenResponse{AccessToken: tok.AccessToken}, nil
}

func (h *BaseOAuthHandler) LoginDefault(req *api.OAuthRequest) (*domain.User, *domain.AuthToken, error) {
	var login string
	var body []byte
	var err error
	user := new(domain.User)
	token := new(domain.AuthToken)
	p := new(domain.Provider)
	up := new(domain.UsersProviders)
	pd := new(domain.ProvidersData)

	if h.Provider == "github" {
		login, body, err = github.GetGitHubUserAndBody(req.AccessToken)
	}
	if h.Provider == "google" {
		login, body, err = google.GetGoogleUserAndBody(req.AccessToken)
	}

	if err != nil {
		return nil, nil, err
	}

	if err = h.providerUsecase.GetByProvider(p, h.Provider); err != nil {
		return nil, nil, err
	}

	h.providersDataUsecase.FindByUsernameAndProvider(pd, login, p.ID)
	if pd.ID == 0 {
		user.Login = login
		if err = h.userUsecase.Create(user); err != nil {
			return nil, nil, err
		}
		if err = h.usersProvidersUsecase.Create(up, user.ID, p.ID); err != nil {
			return nil, nil, err
		}
	}

	if user.ID == "" {
		h.userUsecase.Find(user, pd.UserID)
		if user.ID == "" {
			return nil, nil, errors.New("user not found")
		}
	}

	if pd.ID == 0 {
		pd.UserData = body
		pd.UserID = user.ID
		pd.ProviderID = p.ID
		pd.Username = user.Login
		if err = h.providersDataUsecase.Create(pd); err != nil {
			return nil, nil, err
		}
	}

	tokenStr, err := authtoken.GenerateToken(user.ID, user.Login, h.Provider)
	if err != nil {
		return nil, nil, err
	}

	token.Token = tokenStr
	token.UserUUID = user.ID
	if err = h.tokenUsecase.Create(token); err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

func (h *BaseOAuthHandler) AddAccountDefault(req *api.AddOauthRequest) (*domain.User, error) {
	var login string
	var body []byte
	var err error
	user := new(domain.User)
	up := new(domain.UsersProviders)
	pd := new(domain.ProvidersData)
	p := new(domain.Provider)

	if h.Provider == "github" {
		login, body, err = github.GetGitHubUserAndBody(req.Request.AccessToken)
	}
	if h.Provider == "google" {
		login, body, err = google.GetGoogleUserAndBody(req.Request.AccessToken)
	}

	if err != nil {
		return nil, err
	}

	h.userUsecase.Find(user, req.UserUUID)
	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	h.providerUsecase.GetByProvider(p, h.Provider)
	if p.ID == 0 {
		return nil, errors.New("unknown provider")
	}

	h.providersDataUsecase.FindByUsernameAndProvider(pd, login, p.ID)
	if pd.ID != 0 {
		return nil, errors.New("user already exists")
	}

	if err = h.usersProvidersUsecase.Create(up, req.UserUUID, p.ID); err != nil {
		return nil, err
	}

	pd.UserData = body
	pd.UserID = req.UserUUID
	pd.ProviderID = p.ID
	pd.Username = login
	if err = h.providersDataUsecase.Create(pd); err != nil {
		return nil, err
	}

	return user, nil
}
