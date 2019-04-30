package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"github.com/smallnest/gen/example/model"
)

func configDeptEmpsRouter(router *httprouter.Router) {
	router.GET("/deptemps", GetAllDeptEmps)
	router.POST("/deptemps", AddDeptEmp)
	router.GET("/deptemps/:id", GetDeptEmp)
	router.PUT("/deptemps/:id", UpdateDeptEmp)
	router.DELETE("/deptemps/:id", DeleteDeptEmp)
}

func GetAllDeptEmps(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	deptemps := []*model.DeptEmp{}

	if order != "" {
		err = DB.Model(&model.DeptEmp{}).Order(order).Offset(offset).Limit(pagesize).Find(&deptemps).Error
	} else {
		err = DB.Model(&model.DeptEmp{}).Offset(offset).Limit(pagesize).Find(&deptemps).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deptemp := &model.DeptEmp{}
	if DB.First(deptemp, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, deptemp)
}

func AddDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deptemp := &model.DeptEmp{}

	if err := readJSON(r, deptemp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(deptemp).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, deptemp)
}

func UpdateDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	deptemp := &model.DeptEmp{}
	if DB.First(deptemp, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.DeptEmp{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbmeta.Copy(deptemp, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Save(deptemp).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, deptemp)
}

func DeleteDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deptemp := &model.DeptEmp{}

	if DB.First(deptemp, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(deptemp).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
