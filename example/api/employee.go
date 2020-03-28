package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configEmployeesRouter(router *httprouter.Router) {
	router.GET("/employees", GetAllEmployees)
	router.POST("/employees", AddEmployee)
	router.GET("/employees/:id", GetEmployee)
	router.PUT("/employees/:id", UpdateEmployee)
	router.DELETE("/employees/:id", DeleteEmployee)
}

func configGinEmployeesRouter(router gin.IRoutes) {
	router.GET("/employees", ConverHttprouterToGin(GetAllEmployees))
	router.POST("/employees", ConverHttprouterToGin(AddEmployee))
	router.GET("/employees/:id", ConverHttprouterToGin(GetEmployee))
	router.PUT("/employees/:id", ConverHttprouterToGin(UpdateEmployee))
	router.DELETE("/employees/:id", ConverHttprouterToGin(DeleteEmployee))
}

func GetAllEmployees(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	employees := []*model.Employee{}

	employees_orm := DB.Model(&model.Employee{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		employees_orm = employees_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		employees_orm = employees_orm.Order(order)
	}

	if err = employees_orm.Find(&employees).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
