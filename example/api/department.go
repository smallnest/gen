package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/example/model"
)

func configDepartmentsRouter(router *httprouter.Router) {
	router.GET("/departments", GetAllDepartments)
	router.POST("/departments", PostDepartment)
	router.GET("/departments/:id", GetDepartment)
	router.PUT("/departments/:id", PutDepartment)
	router.DELETE("/departments/:id", DeleteDepartment)
}

func GetAllDepartments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	departments := []model.Department{}
	DB.Find(&departments)
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

func PostDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func PutDepartment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// TODO: copy necessary fields from updated to department

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
