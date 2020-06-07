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
// @Summary Get list of Invoice
// @Tags Invoice
// @Description GetAllInvoice is a handler to get a slice of record(s) from invoices table in the main database
// @Accept  json
// @Produce  json
// @Param   page     query    int     false        "page requested (defaults to 0)"
// @Param   pagesize query    int     false        "number of records in a page  (defaults to 20)"
// @Param   order    query    string  false        "db sort order column"
// @Success 200 {object} api.PagedResults{data=[]model.Invoice}
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices [get]
// http "http://127.0.0.1:8080/invoices?page=0&pagesize=20"
func GetAllInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		returnError(w, r, dao.ErrBadParams)
		return
	}

	pagesize, err := readInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		returnError(w, r, dao.ErrBadParams)
		return
	}

	order := r.FormValue("order")

    records, totalRows, err :=  dao.GetAllInvoices(r.Context(), page, pagesize, order)
	if err != nil {
	    returnError(w, r, err)
		return
	}

	result := &PagedResults{Page: page, PageSize: pagesize, Data: records, TotalRecords: totalRows}
	writeJSON(w, result)
}

```

## Retrieve record
```go

// GetInvoice is a function to get a single record from the invoices table in the main database
// @Summary Get record from table Invoice by  argInvoiceID 
// @Tags Invoice
// @ID argInvoiceID
 // @Description GetInvoice is a function to get a single record from the invoices table in the main database
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
 // @Success 200 {object} model.Invoice
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError "ErrNotFound, db record for id not found - returns NotFound HTTP 404 not found error"
// @Router /invoices/{argInvoiceID} [get]
// http "http://127.0.0.1:8080/invoices/1"
func GetInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {




	argInvoiceID, err := parseInt(ps, "argInvoiceID")
	if err != nil {
		returnError(w, r, err)
		return
	}










	record, err := dao.GetInvoice(r.Context(),  argInvoiceID,        )
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeJSON(w, record)
}

```

## Create record
```go

// AddInvoice add to add a single record to invoices table in the main database
// @Summary Add an record to invoices table
// @Description add to add a single record to invoices table in the main database
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Param Invoice body model.Invoice true "Add Invoice"
// @Success 200 {object} model.Invoice
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices [post]
// echo '{"InvoiceDate": "2267-03-01T02:19:46.038795137-05:00","BillingAddress": "OOBZFjralhvNalrayDTpxLcng","BillingPostalCode": "MIwkbFlJzYyPluFcCWclsywtl","Total": 0.5147250309289902,"InvoiceId": 33,"CustomerId": 81,"BillingCity": "TPSJWkfdaEXTPLmPMwzolIXgU","BillingState": "iWADrSmLOTOxVgBZqfVwLxCeS","BillingCountry": "VhPnkGyZbuNryrFvEQLhWudTd"}' | http POST "http://127.0.0.1:8080/invoices"
func AddInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	invoice := &model.Invoice{}

	if err := readJSON(r, invoice); err != nil {
		returnError(w, r, dao.ErrBadParams)
		return
	}


   if err := invoice.BeforeSave(); err != nil {
      returnError(w, r, dao.ErrBadParams)
   }

   invoice.Prepare()

   if err := invoice.Validate(model.Create); err != nil {
      returnError(w, r, dao.ErrBadParams)
      return
   }

    var err error
	invoice, _, err = dao.AddInvoice(r.Context(), invoice)
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeJSON(w, invoice)
}

```

## Update record
```go

// UpdateInvoice Update a single record from invoices table in the main database
// @Summary Update an record in table invoices
// @Description Update a single record from invoices table in the main database
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
// @Param  Invoice body model.Invoice true "Update Invoice record"
// @Success 200 {object} model.Invoice
// @Failure 400 {object} api.HTTPError
// @Failure 404 {object} api.HTTPError
// @Router /invoices/{argInvoiceID} [patch]
// echo '{"InvoiceDate": "2267-03-01T02:19:46.038795137-05:00","BillingAddress": "OOBZFjralhvNalrayDTpxLcng","BillingPostalCode": "MIwkbFlJzYyPluFcCWclsywtl","Total": 0.5147250309289902,"InvoiceId": 33,"CustomerId": 81,"BillingCity": "TPSJWkfdaEXTPLmPMwzolIXgU","BillingState": "iWADrSmLOTOxVgBZqfVwLxCeS","BillingCountry": "VhPnkGyZbuNryrFvEQLhWudTd"}' | http PUT "http://127.0.0.1:8080/invoices/1"
func UpdateInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {



	argInvoiceID, err := parseInt(ps, "argInvoiceID")
	if err != nil {
		returnError(w, r, err)
		return
	}










	invoice := &model.Invoice{}
	if err := readJSON(r, invoice); err != nil {
		returnError(w, r, dao.ErrBadParams)
		return
	}

   if err := invoice.BeforeSave(); err != nil {
      returnError(w, r, dao.ErrBadParams)
   }

   invoice.Prepare()

   if err := invoice.Validate( model.Update); err != nil {
      returnError(w, r, dao.ErrBadParams)
      return
   }

	invoice, _, err = dao.UpdateInvoice(r.Context(),
	  argInvoiceID,        
	invoice)
	if err != nil {
	    returnError(w, r, err)
   	    return
	}

	writeJSON(w, invoice)
}

```

## Delete record
```go

// DeleteInvoice Delete a single record from invoices table in the main database
// @Summary Delete a record from invoices
// @Description Delete a single record from invoices table in the main database
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Param  argInvoiceID path int true "InvoiceId"
// @Success 204 {object} model.Invoice
// @Failure 400 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Router /invoices/{argInvoiceID} [delete]
// http DELETE "http://127.0.0.1:8080/invoices/1"
func DeleteInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {


	argInvoiceID, err := parseInt(ps, "argInvoiceID")
	if err != nil {
		returnError(w, r, err)
		return
	}










	rowsAffected, err := dao.DeleteInvoice(r.Context(),  argInvoiceID,        )
	if err != nil {
	    returnError(w, r, err)
	    return
	}

	writeRowsAffected(w, rowsAffected )
}

```
