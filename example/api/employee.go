package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"github.com/smallnest/gen/example/model"
)

func configEmployeesRouter(router *httprouter.Router) {
	router.GET("/employees", GetAllEmployees)
	router.POST("/employees", AddEmployee)
	router.GET("/employees/:id", GetEmployee)
	router.PUT("/employees/:id", UpdateEmployee)
	router.DELETE("/employees/:id", DeleteEmployee)
}

func GetAllEmployees(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	employees := []*model.Employee{}

	if order != "" {
		err = DB.Model(&model.Employee{}).Order(order).Offset(offset).Limit(pagesize).Find(&employees).Error
	} else {
		err = DB.Model(&model.Employee{}).Offset(offset).Limit(pagesize).Find(&employees).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func AddEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func UpdateEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	if err := dbmeta.Copy(employee, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
