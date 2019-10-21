package api

import (
	"net/http"

	"generated/model"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configDeptManagersRouter(router *httprouter.Router) {
	router.GET("/deptmanagers", GetAllDeptManagers)
	router.POST("/deptmanagers", AddDeptManager)
	router.GET("/deptmanagers/:id", GetDeptManager)
	router.PUT("/deptmanagers/:id", UpdateDeptManager)
	router.DELETE("/deptmanagers/:id", DeleteDeptManager)
}

func GetAllDeptManagers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	deptmanagers := []*model.DeptManager{}

	if order != "" {
		err = DB.Model(&model.DeptManager{}).Order(order).Offset(offset).Limit(pagesize).Find(&deptmanagers).Error
	} else {
		err = DB.Model(&model.DeptManager{}).Offset(offset).Limit(pagesize).Find(&deptmanagers).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, &deptmanagers)
}

func GetDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deptmanager := &model.DeptManager{}
	if DB.First(deptmanager, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, deptmanager)
}

func AddDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deptmanager := &model.DeptManager{}

	if err := readJSON(r, deptmanager); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(deptmanager).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, deptmanager)
}

func UpdateDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	deptmanager := &model.DeptManager{}
	if DB.First(deptmanager, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.DeptManager{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbmeta.Copy(deptmanager, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Save(deptmanager).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, deptmanager)
}

func DeleteDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deptmanager := &model.DeptManager{}

	if DB.First(deptmanager, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(deptmanager).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
