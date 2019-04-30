package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"github.com/smallnest/gen/example/model"
)

func configDepartmentsRouter(router *httprouter.Router) {
	router.GET("/departments", GetAllDepartments)
	router.POST("/departments", AddDepartment)
	router.GET("/departments/:id", GetDepartment)
	router.PUT("/departments/:id", UpdateDepartment)
	router.DELETE("/departments/:id", DeleteDepartment)
}

func GetAllDepartments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 1)
	if err != nil || page < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	pagesize, err := readInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	offset := (page - 1) * pagesize

	order := r.FormValue("order")

	departments := []*model.Department{}

	if order != "" {
		err = DB.Model(&model.Department{}).Order(order).Offset(offset).Limit(pagesize).Find(&departments).Error
	} else {
		err = DB.Model(&model.Department{}).Offset(offset).Limit(pagesize).Find(&departments).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
