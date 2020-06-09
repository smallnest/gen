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

	invoice = []*model.Invoice{}

	invoiceOrm := DB.Model(&model.Invoice{})
    invoiceOrm.Count(&totalRows)

	if page > 0 {
		offset := (page - 1) * pagesize
		invoiceOrm = invoiceOrm.Offset(offset).Limit(pagesize)
	} else {
		invoiceOrm = invoiceOrm.Limit(pagesize)
    }

	if order != "" {
		invoiceOrm = invoiceOrm.Order(order)
	}

	if err = invoiceOrm.Find(&invoice).Error; err != nil {
	    err = ErrNotFound
		return nil, -1, err
	}

	return invoice, totalRows, nil
}

```

## Retrieve record
```go

// GetInvoice is a function to get a single record from the invoices table in the main database
// error - ErrNotFound, db Find error
func GetInvoice(ctx context.Context,  argInvoiceID int32,        ) (record *model.Invoice, err error) {
	if err = DB.First(&record,   argInvoiceID,        ).Error; err != nil {
	    err = ErrNotFound
		return record, err
	}

	return record, nil
}

```

## Create record
```go

// AddInvoice is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func AddInvoice(ctx context.Context, record *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {
    db := DB.Save(record)
	if err = db.Error; err != nil {
	    return nil, -1, ErrInsertFailed
	}

	return record, db.RowsAffected, nil
}

```

## Update record
```go

// UpdateInvoice is a function to update a single record from invoices table in the main database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateInvoice(ctx context.Context,   argInvoiceID int32,        updated *model.Invoice) (result *model.Invoice, RowsAffected int64, err error) {

   result = &model.Invoice{}
   db := DB.First(result,  argInvoiceID,        )
   if err = db.Error; err != nil {
      return nil, -1, ErrNotFound
   }

   if err = Copy(result, updated); err != nil {
      return nil, -1, ErrUpdateFailed
   }

   db = db.Save(result)
   if err = db.Error; err != nil  {
      return nil, -1, ErrUpdateFailed
   }

   return result, db.RowsAffected, nil
}

```

## Delete record
```go

// DeleteInvoice is a function to delete a single record from invoices table in the main database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteInvoice(ctx context.Context,  argInvoiceID int32,        ) (rowsAffected int64, err error) {

    record := &model.Invoice{}
    db := DB.First(record,   argInvoiceID,        )
    if db.Error != nil {
        return -1, ErrNotFound
    }

    db = db.Delete(record)
    if err = db.Error; err != nil {
        return -1, ErrDeleteFailed
    }

   return db.RowsAffected, nil
}

```
