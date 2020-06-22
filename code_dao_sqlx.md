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

// GetAllInvoices is a function to get a slice of record(s) from invoices table in the main database
// params - page     - page requested (defaults to 0)
// params - pagesize - number of records in a page  (defaults to 20)
// params - order    - db sort order column
// error - ErrNotFound, db Find error
func GetAllInvoices(ctx context.Context, page, pagesize int64, order string) (results []*model.Invoices, totalRows int, err error) {
	sql := "SELECT * FROM `invoices`"

	if order != "" {
		if strings.ContainsAny(order, "'\"") {
			order = ""
		}
	}

	if order == "" {
		order = "InvoiceId"
	}

	if DB.DriverName() == "mssql" {
		sql = fmt.Sprintf("%s order by %s OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", sql, order, page, pagesize)
	} else if DB.DriverName() == "postgres" {
		sql = fmt.Sprintf("%s order by `%s` OFFSET %d LIMIT %d", sql, order, page, pagesize)
	} else {
		sql = fmt.Sprintf("%s order by `%s` LIMIT %d, %d", sql, order, page, pagesize)
	}
	sql = DB.Rebind(sql)

	if Logger != nil {
		Logger(ctx, sql)
	}

	err = DB.SelectContext(ctx, &results, sql)
	if err != nil {
		return nil, -1, err
	}

	cnt , err := GetRowCount(ctx, "invoices")
	if err != nil {
		return results, -2, err
	}

	return results, cnt, err
}


```

## Retrieve record
```go

// GetInvoices is a function to get a single record from the invoices table in the main database
// error - ErrNotFound, db Find error
func GetInvoices(ctx context.Context,  argInvoiceID int32,        ) (record *model.Invoices, err error) {
	sql := "SELECT * FROM `invoices` WHERE InvoiceId = ?"
    sql = DB.Rebind(sql)

    if Logger != nil {
        Logger(ctx, sql)
    }

	record = &model.Invoices{}
	err = DB.GetContext(ctx, record, sql,   argInvoiceID,        )
    if err != nil {
        return nil, err
    }
    return record, nil
}

```

## Create record
```go

// AddInvoices is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func AddInvoices(ctx context.Context, record *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {
	if DB.DriverName() == "postgres" {
		return addInvoicesPostgres( ctx, record)
	} else {
		return addInvoices( ctx, record)
	}
}

// addInvoicesPostgres is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func addInvoicesPostgres(ctx context.Context, record *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {
    sql := "INSERT INTO `invoices` ( CustomerId,  InvoiceDate,  BillingAddress,  BillingCity,  BillingState,  BillingCountry,  BillingPostalCode,  Total) values ( ?, ?, ?, ?, ?, ?, ?, ? )"
    sql = DB.Rebind(sql)

    if Logger != nil {
        Logger(ctx, sql)
    }

    rows := int64(1)
    sql = fmt.Sprintf("%s returning %s", sql, "InvoiceId")
    dbResult := DB.QueryRowContext(ctx, sql,    record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)
    err = dbResult.Scan(   record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)

    return record, rows, err
}

// addInvoicesPostgres is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func addInvoices(ctx context.Context, record *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {
    sql := "INSERT INTO `invoices` ( CustomerId,  InvoiceDate,  BillingAddress,  BillingCity,  BillingState,  BillingCountry,  BillingPostalCode,  Total) values ( ?, ?, ?, ?, ?, ?, ?, ? )"
    sql = DB.Rebind(sql)

    if Logger != nil {
        Logger(ctx, sql)
    }

    rows := int64(0)

    dbResult, err := DB.ExecContext(ctx, sql,    record.CustomerID,  record.InvoiceDate,  record.BillingAddress,  record.BillingCity,  record.BillingState,  record.BillingCountry,  record.BillingPostalCode,  record.Total,)
    if err != nil {
        return nil, 0, err
    }

    id, err := dbResult.LastInsertId()
    rows, err = dbResult.RowsAffected()

     record.InvoiceID = int32(id)


    return record, rows, err
}

```

## Update record
```go

// UpdateInvoices is a function to update a single record from invoices table in the main database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateInvoices(ctx context.Context,   argInvoiceID int32,        updated *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {
	sql := "UPDATE `invoices` set CustomerId = ?, InvoiceDate = ?, BillingAddress = ?, BillingCity = ?, BillingState = ?, BillingCountry = ?, BillingPostalCode = ?, Total = ? WHERE InvoiceId = ?"
	sql = DB.Rebind(sql)

	if Logger != nil {
		Logger(ctx, sql)
	}

	dbResult, err := DB.ExecContext(ctx, sql,    updated.CustomerID,  updated.InvoiceDate,  updated.BillingAddress,  updated.BillingCity,  updated.BillingState,  updated.BillingCountry,  updated.BillingPostalCode,  updated.Total,  argInvoiceID,        )
	if err != nil {
		return nil, 0, err
	}

	rows, err := dbResult.RowsAffected()
      updated.InvoiceID = argInvoiceID
        
	return updated, rows, err
}

```

## Delete record
```go

// DeleteInvoices is a function to delete a single record from invoices table in the main database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteInvoices(ctx context.Context,  argInvoiceID int32,        ) (rowsAffected int64, err error) {
	sql := "DELETE FROM `invoices` where InvoiceId = ?"
	sql = DB.Rebind(sql)

	if Logger != nil {
		Logger(ctx, sql)
	}

	result, err := DB.ExecContext(ctx, sql,   argInvoiceID,        )
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

```
