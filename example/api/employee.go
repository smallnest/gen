package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/example/model"
)

func configEmployeesRouter(router *httprouter.Router) {
	router.GET("/employees", GetAllEmployees)
	router.POST("/employees", PostEmployee)
	router.GET("/employees/:id", GetEmployee)
	router.PUT("/employees/:id", PutEmployee)
	router.DELETE("/employees/:id", DeleteEmployee)
}

func GetAllEmployees(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	employees := []model.Employee{}
	DB.Find(&employees)
	writeJSON(w, &employees)
}

func GetEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	employee := &model.Employee{}
	if DB.First(employee, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, employee)
}

func PostEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	employee := &model.Employee{}

	if err := readJSON(r, employee); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(employee).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, employee)
}

func PutEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	employee := &model.Employee{}
	if DB.First(employee, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.Employee{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: copy necessary fields from updated to employee

	if err := DB.Save(employee).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, employee)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	employee := &model.Employee{}

	if DB.First(employee, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(employee).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
