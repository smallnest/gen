package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configDeptEmpsRouter(router *httprouter.Router) {
	router.GET("/deptemps", GetAllDeptEmps)
	router.POST("/deptemps", AddDeptEmp)
	router.GET("/deptemps/:id", GetDeptEmp)
	router.PUT("/deptemps/:id", UpdateDeptEmp)
	router.DELETE("/deptemps/:id", DeleteDeptEmp)
}

func configGinDeptEmpsRouter(router gin.IRoutes) {
	router.GET("/deptemps", ConverHttprouterToGin(GetAllDeptEmps))
	router.POST("/deptemps", ConverHttprouterToGin(AddDeptEmp))
	router.GET("/deptemps/:id", ConverHttprouterToGin(GetDeptEmp))
	router.PUT("/deptemps/:id", ConverHttprouterToGin(UpdateDeptEmp))
	router.DELETE("/deptemps/:id", ConverHttprouterToGin(DeleteDeptEmp))
}

func GetAllDeptEmps(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	deptemps := []*model.DeptEmp{}

	deptemps_orm := DB.Model(&model.DeptEmp{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		deptemps_orm = deptemps_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		deptemps_orm = deptemps_orm.Order(order)
	}

	if err = deptemps_orm.Find(&deptemps).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, &deptemps)
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
