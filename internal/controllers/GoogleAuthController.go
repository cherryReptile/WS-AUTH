package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/pavel-one/GoStarter/api"
	"github.com/pavel-one/GoStarter/internal/helpers"
	"github.com/pavel-one/GoStarter/internal/resources/requests"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"os"
	"strings"
)

type GoogleAuthController struct {
	BaseOAuthController
	GoogleService api.AuthGoogleServiceClient
	Config        *oauth2.Config
}

func (c *GoogleAuthController) Init(gs api.AuthGoogleServiceClient) {
	c.GoogleService = gs
	c.Config = &oauth2.Config{}
	c.Config.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	c.Config.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	c.Config.Scopes = []string{"https://www.googleapis.com/auth/userinfo.email"}
	c.Config.Endpoint = google.Endpoint
}

var GoogleRedirectLogin = "/api/v1/auth/google/login"
var GoogleRedirectToExchangeToken = "/api/v1/auth/google/token"

func (c *GoogleAuthController) RedirectForAuth(ctx *gin.Context) {
	c.Config.RedirectURL = "http://" + "localhost" + GoogleRedirectToExchangeToken
	u := c.Config.AuthCodeURL(c.setOAuthStateCookie(ctx, GoogleRedirectToExchangeToken, "localhost"))
	ctx.Redirect(http.StatusTemporaryRedirect, u)
}

func (c *GoogleAuthController) GetAccessToken(ctx *gin.Context) {
	code, err := c.checkOAuthStateCookie(ctx)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	tok, err := c.Config.Exchange(context.Background(), code)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"google_access_token": tok.AccessToken})
}

func (c *GoogleAuthController) Login(ctx *gin.Context) {
	t := new(requests.Token)
	if err := ctx.ShouldBindJSON(t); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	login, body, err := c.getGoogleUserAndBody(t.Token)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	res, err := c.GoogleService.Login(context.Background(), &api.GoogleRequest{Email: login, Data: body})
	if err != nil {
		e := strings.Split(err.Error(), "=")
		c.ERROR(ctx, http.StatusBadRequest, errors.New(e[2]))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": res.Struct, "token": res.TokenStr})
}

func (c *GoogleAuthController) getGoogleUserAndBody(token string) (string, []byte, error) {
	user := new(requests.GoogleUser)
	res, err := helpers.RequestToGoogle(token)
	if err != nil {
		return "", nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}

	err = json.Unmarshal(body, user)

	if err != nil {
		return "", nil, err
	}

	return user.Email, body, nil
}
