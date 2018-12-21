package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/example/model"
)

func configDeptEmpsRouter(router *httprouter.Router) {
	router.GET("/deptemps", GetAllDeptEmps)
	router.POST("/deptemps", PostDeptEmp)
	router.GET("/deptemps/:id", GetDeptEmp)
	router.PUT("/deptemps/:id", PutDeptEmp)
	router.DELETE("/deptemps/:id", DeleteDeptEmp)
}

func GetAllDeptEmps(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deptemps := []model.DeptEmp{}
	DB.Find(&deptemps)
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

func PostDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func PutDeptEmp(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// TODO: copy necessary fields from updated to deptemp

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
