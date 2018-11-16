package controllers

import (
	"fmt"

	"github.com/mataharibiz/sange"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"

	"midas/models/models"
)

// RfqController Rfq  controller
type RfqController struct {
	DB *gorm.DB
}

// COPY BELOW CODE TO THE ROUTE CONFIG
// app.PartyFunc("/rfq", func(rfq iris.Party) {
// 	rfqController := &controllers.RfqsController{DB: r.DB}
// 	rfqController.Get("/", rfqController.GetAllRfqs)
// 	rfqController.Post("/", rfqController.AddRfq)
// 	rfqController.Get("/{id}", rfqController.GetRfq)
// 	rfqController.Put("/{id}", rfqController.UpdateRfq)
// 	rfqController.Delete("/{id}", rfqController.DeleteRfq)
// })

// GetAllRfqs controller for get data list
func (controller *RfqsController) GetAllRfqs(ctx iris.Context) {
	var err error

	// Get page number
	page := ctx.URLParamIntDefault("page", 1)
	if page < 1 {
		sange.NewResponse(ctx, iris.StatusBadRequest, fmt.Errorf("Parameter [page] is not valid"))
		return
	}

	// Get Limit
	pagesize := ctx.URLParamIntDefault("limit", 1000)
	if pagesize <= 0 {
		sange.NewResponse(ctx, iris.StatusBadRequest, fmt.Errorf("Parameter [limit] is not valid"))
		return
	}
	offset := (page - 1) * pagesize

	// Get Order By
	order := ctx.URLParamDefault("order", "id ASC")

	rfqs := []*model.Rfq{}

	if order != "" {
		err = controller.DB.Model(&models.Rfq{}).Order(order).Offset(offset).Limit(pagesize).Find(&rfqs).Error
	} else {
		err = controller.DB.Model(&models.Rfq{}).Offset(offset).Limit(pagesize).Find(&rfqs).Error
	}

	if err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{
		"items":      rfqs,
		"pagination": map[string]interface{}{}, //TODO: add pagination logic
	}

	sange.NewResponse(ctx, iris.StatusOK, resp)
}

// GetRfq get single data
func (controller *RfqsController) GetRfq(ctx iris.Context) {

	id := ctx.Params().Get("id")
	rfq := &models.Rfq{}

	result := controller.DB.Find(rfq, id)

	if result.RecordNotFound() {
		rfq = nil
	}

	sange.NewResponse(ctx, iris.StatusOK, rfq)
}

// AddRfq add or save data
func (controller *RfqsController) AddRfq(ctx iris.Context) {
	rfq := &models.Rfq{}

	if err := ctx.ReadJSON(invoice); err != nil {
		sange.NewResponse(ctx, iris.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := controller.DB.Save(rfq).Error; err != nil {
		sange.NewResponse(ctx, iris.StatusUnprocessableEntity, err.Error())
		return
	}

	sange.NewResponse(ctx, iris.StatusOK, rfq)
}

// UpdateRfq update data
func (controller *RfqsController) UpdateRfq(ctx iris.Context) {
	id := ctx.Params().Get("id")

	rfq := &models.Rfq{}
	result := controller.DB.First(rfq, id)
	if result.RecordNotFound() {
		sange.NewResponse(ctx, iris.StatusNotFound, "rfq does not exist")
		return
	}

	updated := &models.Rfq{}
	if err := ctx.ReadJSON(updated); err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	if err := controller.DB.Update(updated).Error; err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	sange.NewResponse(ctx, iris.StatusOK, updated)
}

// UpdateRfq delete data
func (controller *RfqsController) DeleteRfq(ctx iris.Context) {
	id := ctx.Params().Get("id")
	rfq := &models.Rfq{}

	result := controller.DB.First(rfq, id)
	if result.RecordNotFound() {
		sange.NewResponse(ctx, iris.StatusNotFound, "rfq does not exist")
		return
	}
	if err := controller.DB.Delete(rfq).Error; err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	sange.NewResponse(ctx, iris.StatusOK, "Delete successfull")
}
