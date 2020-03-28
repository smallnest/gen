package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
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

func configGinDeptManagersRouter(router gin.IRoutes) {
	router.GET("/deptmanagers", ConverHttprouterToGin(GetAllDeptManagers))
	router.POST("/deptmanagers", ConverHttprouterToGin(AddDeptManager))
	router.GET("/deptmanagers/:id", ConverHttprouterToGin(GetDeptManager))
	router.PUT("/deptmanagers/:id", ConverHttprouterToGin(UpdateDeptManager))
	router.DELETE("/deptmanagers/:id", ConverHttprouterToGin(DeleteDeptManager))
}

func GetAllDeptManagers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	deptmanagers := []*model.DeptManager{}

	deptmanagers_orm := DB.Model(&model.DeptManager{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		deptmanagers_orm = deptmanagers_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		deptmanagers_orm = deptmanagers_orm.Order(order)
	}

	if err = deptmanagers_orm.Find(&deptmanagers).Error; err != nil {
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
