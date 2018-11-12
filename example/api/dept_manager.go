package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nqsang90/gen/example/model"
)

func configDeptManagersRouter(router *httprouter.Router) {
	router.GET("/deptmanagers", GetAllDeptManagers)
	router.POST("/deptmanagers", PostDeptManager)
	router.GET("/deptmanagers/:id", GetDeptManager)
	router.PUT("/deptmanagers/:id", PutDeptManager)
	router.DELETE("/deptmanagers/:id", DeleteDeptManager)
}

func GetAllDeptManagers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deptmanagers := []model.DeptManager{}
	DB.Find(&deptmanagers)
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

func PostDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func PutDeptManager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// TODO: copy necessary fields from updated to deptmanager

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
