package template

var ControllerTmpl = `package controllers

import (
	"fmt"

	"github.com/mataharibiz/sange"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"

	"{{.PackageName}}/models"
)

// {{.StructName}}Controller {{.StructName}}  controller
type {{.StructName}}Controller struct {
	DB *gorm.DB
}

// COPY BELOW CODE TO THE ROUTE CONFIG
// app.PartyFunc("/{{.StructName | toSnakeCase}}", func({{.StructName | toLowerCamelCase}} iris.Party) {
// 	{{.StructName | toLowerCamelCase}}Controller := &controllers.{{pluralize .StructName}}Controller{DB: r.DB}
// 	{{.StructName | toLowerCamelCase}}Controller.Get("/", {{.StructName | toLowerCamelCase}}Controller.GetAll{{pluralize .StructName}})
// 	{{.StructName | toLowerCamelCase}}Controller.Post("/", {{.StructName | toLowerCamelCase}}Controller.Add{{.StructName}})
// 	{{.StructName | toLowerCamelCase}}Controller.Get("/{id}", {{.StructName | toLowerCamelCase}}Controller.Get{{.StructName}})
// 	{{.StructName | toLowerCamelCase}}Controller.Put("/{id}", {{.StructName | toLowerCamelCase}}Controller.Update{{.StructName}})
// 	{{.StructName | toLowerCamelCase}}Controller.Delete("/{id}", {{.StructName | toLowerCamelCase}}Controller.Delete{{.StructName}})
// })

// GetAll{{pluralize .StructName}} controller for get data list
func (controller *{{pluralize .StructName}}Controller) GetAll{{pluralize .StructName}}(ctx iris.Context) {
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

	{{pluralize .StructName | toLowerCamelCase}} := []*model.{{.StructName}}{}

	if order != "" {
		err = controller.DB.Model(&models.{{.StructName}}{}).Order(order).Offset(offset).Limit(pagesize).Find(&{{pluralize .StructName | toLowerCamelCase}}).Error
	} else {
		err = controller.DB.Model(&models.{{.StructName}}{}).Offset(offset).Limit(pagesize).Find(&{{pluralize .StructName | toLowerCamelCase}}).Error
	}

	if err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{
		"items":      {{pluralize .StructName | toLowerCamelCase}},
		"pagination": map[string]interface{}{}, //TODO: add pagination logic
	}

	sange.NewResponse(ctx, iris.StatusOK, resp)
}

// Get{{.StructName}} get single data
func (controller *{{pluralize .StructName}}Controller) Get{{.StructName}}(ctx iris.Context) {

	id := ctx.Params().Get("id")
	{{.StructName | toLowerCamelCase}} := &models.{{.StructName}}{}

	result := controller.DB.Find({{.StructName | toLower}}, id)

	if result.RecordNotFound() {
		{{.StructName | toLowerCamelCase}} = nil
	}

	sange.NewResponse(ctx, iris.StatusOK, {{.StructName | toLowerCamelCase}})
}

// Add{{.StructName}} add or save data
func (controller *{{pluralize .StructName}}Controller) Add{{.StructName}}(ctx iris.Context) {
	{{.StructName | toLowerCamelCase}} := &models.{{.StructName}}{}

	if err := ctx.ReadJSON(invoice); err != nil {
		sange.NewResponse(ctx, iris.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := controller.DB.Save({{.StructName | toLowerCamelCase}}).Error; err != nil {
		sange.NewResponse(ctx, iris.StatusUnprocessableEntity, err.Error())
		return
	}

	sange.NewResponse(ctx, iris.StatusOK, {{.StructName | toLowerCamelCase}})
}

// Update{{.StructName}} update data
func (controller *{{pluralize .StructName}}Controller) Update{{.StructName}}(ctx iris.Context) {
	id := ctx.Params().Get("id")

	{{.StructName | toLowerCamelCase}} := &models.{{.StructName}}{}
	result := controller.DB.First({{.StructName | toLowerCamelCase}}, id)
	if result.RecordNotFound() {
		sange.NewResponse(ctx, iris.StatusNotFound, "{{.StructName | toSnakeCase}} does not exist")
		return
	}

	updated := &models.{{.StructName}}{}
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

// Update{{.StructName}} delete data
func (controller *{{pluralize .StructName}}Controller) Delete{{.StructName}}(ctx iris.Context) {
	id := ctx.Params().Get("id")
	{{.StructName | toLowerCamelCase}} := &models.{{.StructName}}{}

	result := controller.DB.First({{.StructName | toLowerCamelCase}}, id)
	if result.RecordNotFound() {
		sange.NewResponse(ctx, iris.StatusNotFound, "{{.StructName | toSnakeCase}} does not exist")
		return
	}
	if err := controller.DB.Delete({{.StructName | toLowerCamelCase}}).Error; err != nil {
		sange.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	sange.NewResponse(ctx, iris.StatusOK, "Delete successfull")
}

`
