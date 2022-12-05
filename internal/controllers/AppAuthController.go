package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-one/GoStarter/internal/appauth"
	"github.com/pavel-one/GoStarter/internal/models"
	"github.com/pavel-one/GoStarter/internal/resources/requests"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

type AppAuthController struct {
	BaseJwtAuthController
}

func (c *AppAuthController) Init(db *sqlx.DB) {
	c.DB = db
}

func (c *AppAuthController) Register(ctx *gin.Context) {
	user := new(models.User)
	tokenModel := new(models.AccessToken)
	reqU := new(requests.UserRequest)
	if err := ctx.ShouldBindJSON(reqU); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	user.FindByUniqueAndService(c.DB, reqU.Email, "app")
	if user.ID != 0 {
		c.ERROR(ctx, http.StatusBadRequest, errors.New("this user already exists"))
		return
	}

	hashP, err := bcrypt.GenerateFromPassword([]byte(reqU.Password), bcrypt.DefaultCost)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	user.Password = string(hashP)
	user.UniqueRaw = reqU.Email
	user.AuthorizedBy = "app"

	err = user.Create(c.DB)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	if user.ID == 0 {
		c.ERROR(ctx, http.StatusBadRequest, errors.New("user not found"))
		return
	}

	tokenStr, err := appauth.GenerateToken(user.ID, user.UniqueRaw, user.AuthorizedBy)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	tokenModel.Token = tokenStr
	tokenModel.UserID = user.ID

	if err = tokenModel.Create(c.DB); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	//c.setServiceCookie(ctx, "app", os.Getenv("DOMAIN"))
	//c.setUIDCookie(ctx, user.Email, os.Getenv("DOMAIN"))

	ctx.JSON(http.StatusOK, gin.H{"user": user, "token": tokenModel})
}

func (c *AppAuthController) Login(ctx *gin.Context) {
	user := new(models.User)
	tokenModel := new(models.AccessToken)

	reqU := new(requests.UserRequest)
	if err := ctx.ShouldBindJSON(reqU); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	user.UniqueRaw = reqU.Email
	if err := user.FindByUniqueAndService(c.DB, user.UniqueRaw, "app"); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	if user.ID == 0 {
		c.ERROR(ctx, http.StatusBadRequest, errors.New("user not found"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqU.Password)); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	tokenStr, err := appauth.GenerateToken(user.ID, user.UniqueRaw, "app")
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	tokenModel.Token = tokenStr
	tokenModel.UserID = user.ID
	if err = tokenModel.Create(c.DB); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	c.setServiceCookie(ctx, "app", os.Getenv("DOMAIN"))

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusText(http.StatusOK), "token": tokenModel.Token})
}

func (c *AppAuthController) Logout(ctx *gin.Context) {
	if err := c.LogoutFromApp(ctx, c.DB); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logout successfully"})
}
