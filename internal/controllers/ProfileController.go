package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/pavel-one/GoStarter/api"
	"github.com/pavel-one/GoStarter/internal/helpers"
	"github.com/pavel-one/GoStarter/internal/resources/requests"
	"net/http"
)

type ProfileController struct {
	BaseController
	ProfileService api.ProfileServiceClient
}

func (c *ProfileController) Init(ps api.ProfileServiceClient) {
	c.ProfileService = ps
}

func (c *ProfileController) Create(ctx *gin.Context) {
	p := new(requests.Profile)
	if err := ctx.ShouldBindJSON(p); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	if err := c.checkRequest(p); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	od, err := json.Marshal(map[string]string{"data": p.OtherData})
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	uuid, err := helpers.GetAndCastUserUUID(ctx)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	res, err := c.ProfileService.Create(context.Background(), &api.ProfileRequest{
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Other_Data: od,
		Address:    p.Address,
		UserUUID:   uuid,
	})

	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusCreated, c.toResponse(res))
}

func (c *ProfileController) Get(ctx *gin.Context) {
	uuid, err := helpers.GetAndCastUserUUID(ctx)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	res, err := c.ProfileService.Get(context.Background(), &api.ProfileUUID{UserUUID: uuid})
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, c.toResponse(res))
}

func (c *ProfileController) Update(ctx *gin.Context) {
	p := new(requests.Profile)
	if err := ctx.ShouldBindJSON(p); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	if err := c.checkRequest(p); err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	od, err := json.Marshal(map[string]string{"data": p.OtherData})
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	uuid, err := helpers.GetAndCastUserUUID(ctx)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	res, err := c.ProfileService.Update(context.Background(), &api.ProfileRequest{
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Other_Data: od,
		Address:    p.Address,
		UserUUID:   uuid,
	})

	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, c.toResponse(res))
}

func (c *ProfileController) Delete(ctx *gin.Context) {
	uuid, err := helpers.GetAndCastUserUUID(ctx)
	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	res, err := c.ProfileService.Delete(context.Background(), &api.ProfileUUID{UserUUID: uuid})

	if err != nil {
		c.ERROR(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": res.Message})
}

func (c *ProfileController) checkRequest(p *requests.Profile) error {
	if p.FirstName == "" && p.LastName == "" && p.OtherData == "" && p.Address == "" {
		return errors.New("all fields are required")
	}

	return nil
}

func (c *ProfileController) toResponse(res *api.ProfileResponse) gin.H {
	return gin.H{
		"first_name": res.FirstName,
		"last_name":  res.LastName,
		"other_data": res.Other_Data,
		"address":    res.Address,
	}
}
