[comment]: <> (This is a generated file please edit source in ./templates)
[comment]: <> (All modification will be lost, you have been warned)
[comment]: <> ()

## CRUD Http Handlers
`gen` will generate http handlers if the `--rest` is used. The code can be customized with the `--api=api` flag to set the name of the api package.

- [Retrieving records with paging](#Retrieve-Paged-Records)
- [Retrieve a specific record](#Retrieve-record)
- [Create a record](#Create-record)
- [Update a record](#Update-record)
- [Delete a record](#Delete-record)

`gen` will add swagger comments to the source generated, this too can be customized with the following.
```bash
  --swagger_version=1.0                           swagger version
  --swagger_path=/                                swagger base path
  --swagger_tos=                                  swagger tos url
  --swagger_contact_name=Me                       swagger contact name
  --swagger_contact_url=http://me.com/terms.html  swagger contact url
  --swagger_contact_email=me@me.com               swagger contact email
```


## Retrieve Paged Records
```go

// GetAllInvoices is a function to get a slice of record(s) from invoices table in the main database
// @Summary Get list of Invoices
// @Tags Invoices
// @Description GetAllInvoices is a handler to get a slice of record(s) from invoices table in the main database
// @Accept  json
// @Produce  json
// @Param   page     query    int     false        "page requested (defaults to 0)"
// @Param   pagesize query    int     false        "number of records in a page  (defaults to 20)"
// @Param   order    query    string  false        "db sort order column"
// @Success 200 {object} api.PagedResults{data=[]model.Invoices}
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices [get]
// http "http://127.0.0.1:8080/invoices?page=0&pagesize=20" X-Api-User:user123
func GetAllInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := initializeContext(r)
    page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		returnError(ctx, w, r, dao.ErrBadParams)
		return
	}

	pagesize, err := readInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		returnError(ctx, w, r, dao.ErrBadParams)
		return
	}

	order := r.FormValue("order")

	if err := ValidateRequest(ctx, r, "invoices", model.RetrieveMany); err != nil{
		returnError(ctx, w, r, err)
		return
	}

    records, totalRows, err :=  dao.GetAllInvoices(ctx, page, pagesize, order)
	if err != nil {
	    returnError(ctx, w, r, err)
		return
	}

	result := &PagedResults{Page: page, PageSize: pagesize, Data: records, TotalRecords: totalRows}
	writeJSON(ctx, w, result)
}

```

## Retrieve record
```go

// GetInvoices is a function to get a single record from the invoices table in the main database
// @Summary Get record from table Invoices by  argInvoiceID 
// @Tags Invoices
// @ID argInvoiceID
 // @Description GetInvoices is a function to get a single record from the invoices table in the main database
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
 // @Success 200 {object} model.Invoices
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError "ErrNotFound, db record for id not found - returns NotFound HTTP 404 not found error"
// @Router /invoices/{argInvoiceID} [get]
// http "http://127.0.0.1:8080/invoices/1" X-Api-User:user123
func GetInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := initializeContext(r)



	argInvoiceID, err := parseInt32(ps, "argInvoiceID")
	if err != nil {
		returnError(ctx, w, r, err)
		return
	}










	if err := ValidateRequest(ctx, r, "invoices", model.RetrieveOne); err != nil{
		returnError(ctx, w, r, err)
		return
	}

	record, err := dao.GetInvoices(ctx,  argInvoiceID,        )
	if err != nil {
		returnError(ctx, w, r, err)
		return
	}

	writeJSON(ctx, w, record)
}

```

## Create record
```go

// AddInvoices add to add a single record to invoices table in the main database
// @Summary Add an record to invoices table
// @Description add to add a single record to invoices table in the main database
// @Tags Invoices
// @Accept  json
// @Produce  json
// @Param Invoices body model.Invoices true "Add Invoices"
// @Success 200 {object} model.Invoices
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices [post]
// echo '{"billing_state": "ZDIUXoQexYEYgpaYJgyitZpjS","billing_country": "xbuWQjVGOCsrKSXCJhxeVSFhQ","billing_postal_code": "fEHjgTFVSOwVZEPNUjpVkmOZV","customer_id": 40,"invoice_date": "2021-12-17T01:29:16.603984343-08:00","billing_address": "GoPOJOqToKgHAXoCRwOLzSsgW","billing_city": "pRhdTdgdcpgiBrgvmCMlFKapm","total": 0.38248355155004915,"invoice_id": 46}' | http POST "http://127.0.0.1:8080/invoices" X-Api-User:user123
func AddInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := initializeContext(r)
	invoices := &model.Invoices{}

	if err := readJSON(r, invoices); err != nil {
		returnError(ctx, w, r, dao.ErrBadParams)
		return
	}


   if err := invoices.BeforeSave(); err != nil {
      returnError(ctx, w, r, dao.ErrBadParams)
   }

   invoices.Prepare()

   if err := invoices.Validate(model.Create); err != nil {
      returnError(ctx, w, r, dao.ErrBadParams)
      return
   }

	if err := ValidateRequest(ctx, r, "invoices", model.Create); err != nil{
		returnError(ctx, w, r, err)
		return
	}

    var err error
	invoices, _, err = dao.AddInvoices(ctx, invoices)
	if err != nil {
		returnError(ctx, w, r, err)
		return
	}

	writeJSON(ctx, w, invoices)
}

```

## Update record
```go

// UpdateInvoices Update a single record from invoices table in the main database
// @Summary Update an record in table invoices
// @Description Update a single record from invoices table in the main database
// @Tags Invoices
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
// @Param  Invoices body model.Invoices true "Update Invoices record"
// @Success 200 {object} model.Invoices
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices/{argInvoiceID} [put]
// echo '{"billing_state": "ZDIUXoQexYEYgpaYJgyitZpjS","billing_country": "xbuWQjVGOCsrKSXCJhxeVSFhQ","billing_postal_code": "fEHjgTFVSOwVZEPNUjpVkmOZV","customer_id": 40,"invoice_date": "2021-12-17T01:29:16.603984343-08:00","billing_address": "GoPOJOqToKgHAXoCRwOLzSsgW","billing_city": "pRhdTdgdcpgiBrgvmCMlFKapm","total": 0.38248355155004915,"invoice_id": 46}' | http PUT "http://127.0.0.1:8080/invoices/1"  X-Api-User:user123
func UpdateInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := initializeContext(r)



	argInvoiceID, err := parseInt32(ps, "argInvoiceID")
	if err != nil {
		returnError(ctx, w, r, err)
		return
	}










	invoices := &model.Invoices{}
	if err := readJSON(r, invoices); err != nil {
		returnError(ctx, w, r, dao.ErrBadParams)
		return
	}

   if err := invoices.BeforeSave(); err != nil {
      returnError(ctx, w, r, dao.ErrBadParams)
   }

   invoices.Prepare()

   if err := invoices.Validate( model.Update); err != nil {
      returnError(ctx, w, r, dao.ErrBadParams)
      return
   }

	if err := ValidateRequest(ctx, r, "invoices", model.Update); err != nil{
		returnError(ctx, w, r, err)
		return
	}

	invoices, _, err = dao.UpdateInvoices(ctx,
	  argInvoiceID,        
	invoices)
	if err != nil {
	    returnError(ctx, w, r, err)
   	    return
	}

	writeJSON(ctx, w, invoices)
}

```

## Delete record
```go

// DeleteInvoices Delete a single record from invoices table in the main database
// @Summary Delete a record from invoices
// @Description Delete a single record from invoices table in the main database
// @Tags Invoices
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
// @Success 204 {object} model.Invoices
// @Failure 400 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Router /invoices/{argInvoiceID} [delete]
// http DELETE "http://127.0.0.1:8080/invoices/1" X-Api-User:user123
func DeleteInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := initializeContext(r)


	argInvoiceID, err := parseInt32(ps, "argInvoiceID")
	if err != nil {
		returnError(ctx, w, r, err)
		return
	}










	if err := ValidateRequest(ctx, r, "invoices", model.Delete); err != nil{
		returnError(ctx, w, r, err)
		return
	}

	rowsAffected, err := dao.DeleteInvoices(ctx,  argInvoiceID,        )
	if err != nil {
	    returnError(ctx, w, r, err)
	    return
	}

	writeRowsAffected(w, rowsAffected )
}

```
