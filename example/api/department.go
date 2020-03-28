package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configDepartmentsRouter(router *httprouter.Router) {
	router.GET("/departments", GetAllDepartments)
	router.POST("/departments", AddDepartment)
	router.GET("/departments/:id", GetDepartment)
	router.PUT("/departments/:id", UpdateDepartment)
	router.DELETE("/departments/:id", DeleteDepartment)
}

func configGinDepartmentsRouter(router gin.IRoutes) {
	router.GET("/departments", ConverHttprouterToGin(GetAllDepartments))
	router.POST("/departments", ConverHttprouterToGin(AddDepartment))
	router.GET("/departments/:id", ConverHttprouterToGin(GetDepartment))
	router.PUT("/departments/:id", ConverHttprouterToGin(UpdateDepartment))
	router.DELETE("/departments/:id", ConverHttprouterToGin(DeleteDepartment))
}

func GetAllDepartments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	departments := []*model.Department{}

	departments_orm := DB.Model(&model.Department{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		departments_orm = departments_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		departments_orm = departments_orm.Order(order)
	}

	if err = departments_orm.Find(&departments).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, &departments)
}

func GetDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	department := &model.Department{}
	if DB.First(department, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, department)
}

func AddDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	department := &model.Department{}

	if err := readJSON(r, department); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(department).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, department)
}

func UpdateDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	department := &model.Department{}
	if DB.First(department, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.Department{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbmeta.Copy(department, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Save(department).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, department)
}

func DeleteDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	department := &model.Department{}

	if DB.First(department, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(department).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
