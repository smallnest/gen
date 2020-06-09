[comment]: <> (This is a generated file please edit source in ./templates)
[comment]: <> (All modification will be lost, you have been warned)
[comment]: <> ()

## CRUD DAO Functions
`gen` will generate dao functions if the `--generate-dao` is passed to `gen`. The code can be customized with the `--dao=dao` flag to set the name of the dao package.

Code can be generated in two flavours, SQLX by default and GORM with the flag `--gorm`


The code generation, will generate functions for
- [Retrieving records with paging](#Retrieve-Paged-Records)
- [Retrieve a specific record](#Retrieve-record)
- [Create a record](#Create-record)
- [Update a record](#Update-record)
- [Delete a record](#Delete-record)

## Retrieve Paged Records
```go

// GetAllInvoice is a function to get a slice of record(s) from invoices table in the main database
// params - page     - page requested (defaults to 0)
// params - pagesize - number of records in a page  (defaults to 20)
// params - order    - db sort order column
// error - ErrNotFound, db Find error
func GetAllInvoice(ctx context.Context, page, pagesize int64, order string) (invoice []*model.Invoice, totalRows int, err error) {
	sql := "SELECT * FROM invoices"

	if order == "" {
	    order = "InvoiceId"
	}

	if DB.DriverName() == "mssql" {
		sql = fmt.Sprintf("%s order by %s OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", sql, order, page, pagesize)
	} else if DB.DriverName() == "postgres" {
		sql = fmt.Sprintf("%s order by %s OFFSET %d LIMIT %d", sql, order, page, pagesize)
	} else {
		sql = fmt.Sprintf("%s order by %s LIMIT %d, %d", sql, order, page, pagesize)
	}

	err = DB.SelectContext(ctx, &invoice, sql)
	return invoice, len(invoice), err
}

```

## Retrieve record
```go

// GetInvoice is a function to get a single record from the invoices table in the main database
// error - ErrNotFound, db Find error
func GetInvoice(ctx context.Context,  argInvoiceID int32,        ) (record *model.Invoice, err error) {
	sql := "SELECT * FROM invoices WHERE InvoiceId = $1"
	record = &model.Invoice{}
	err = DB.GetContext(ctx, record, sql,   argInvoiceID,        )
    if err != nil {
        return nil, err
    }
    return record, nil
}

```

## Create record
```go

// AddInvoice is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func AddInvoice(ctx context.Context, record *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {
	if DB.DriverName() == "postgres" {
		return addInvoicePostgres( ctx, record)
	} else {
		return addInvoice( ctx, record)
	}
}

// addInvoicePostgres is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func addInvoicePostgres(ctx context.Context, record *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {
    sql := "INSERT INTO invoices ( CustomerId,  InvoiceDate,  BillingAddress,  BillingCity,  BillingState,  BillingCountry,  BillingPostalCode,  Total) values ( $1, $2, $3, $4, $5, $6, $7, $8 )"

    rows := int64(1)
    sql = fmt.Sprintf("%s returning %s", sql, "InvoiceId")
    dbResult := DB.QueryRowContext(ctx, sql,    record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)
    err = dbResult.Scan(   record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)

    return record, rows, err
}

// addInvoicePostgres is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func addInvoice(ctx context.Context, record *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {
    sql := "INSERT INTO invoices ( CustomerId,  InvoiceDate,  BillingAddress,  BillingCity,  BillingState,  BillingCountry,  BillingPostalCode,  Total) values ( $1, $2, $3, $4, $5, $6, $7, $8 )"

    rows := int64(0)

    dbResult := DB.MustExecContext(ctx, sql,    record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)
    id, err := dbResult.LastInsertId()
    rows, err = dbResult.RowsAffected()

     record.InvoiceID = int32(id)



    return record, rows, err
}

```

## Update record
```go

// UpdateInvoice is a function to update a single record from invoices table in the main database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateInvoice(ctx context.Context,   argInvoiceID int32,        updated *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {
	sql := "UPDATE invoices set CustomerId = $1, InvoiceDate = $2, BillingAddress = $3, BillingCity = $4, BillingState = $5, BillingCountry = $6, BillingPostalCode = $7, Total = $8 WHERE InvoiceId = $9"
	dbResult := DB.MustExecContext(ctx, sql,    updated.CustomerID,  updated.InvoiceDate,  updated.BillingAddress,  updated.BillingCity,  updated.BillingState,  updated.BillingCountry,  updated.BillingPostalCode,  updated.Total,  argInvoiceID,        )
	rows, err := dbResult.RowsAffected()
      updated.InvoiceID = argInvoiceID
        
	return updated, rows, err
}

```

## Delete record
```go

// DeleteInvoice is a function to delete a single record from invoices table in the main database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteInvoice(ctx context.Context,  argInvoiceID int32,        ) (rowsAffected int64, err error) {
	sql := "DELETE FROM invoices where InvoiceId = $1"
	result := DB.MustExecContext(ctx, sql,   argInvoiceID,        )
	return result.RowsAffected()
}

```
