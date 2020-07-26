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

	resultOrm := DB.Model(&model.Invoices{})
    resultOrm.Count(&totalRows)

	if page > 0 {
		offset := (page - 1) * pagesize
		resultOrm = resultOrm.Offset(offset).Limit(pagesize)
	} else {
		resultOrm = resultOrm.Limit(pagesize)
    }

	if order != "" {
		resultOrm = resultOrm.Order(order)
	}

	if err = resultOrm.Find(&results).Error; err != nil {
	    err = ErrNotFound
		return nil, -1, err
	}

	return results, totalRows, nil
}

```

## Retrieve record
```go

// GetInvoices is a function to get a single record from the invoices table in the main database
// error - ErrNotFound, db Find error
func GetInvoices(ctx context.Context,  argInvoiceID int32,        ) (record *model.Invoices, err error) {
	record = &model.Invoices{}
	if err = DB.First(record,   argInvoiceID,        ).Error; err != nil {
	    err = ErrNotFound
		return record, err
	}

	return record, nil
}

```

## Create record
```go

// AddInvoices is a function to add a single record to invoices table in the main database
// error - ErrInsertFailed, db save call failed
func AddInvoices(ctx context.Context, record *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {
    db := DB.Save(record)
	if err = db.Error; err != nil {
	    return nil, -1, ErrInsertFailed
	}

	return record, db.RowsAffected, nil
}

```

## Update record
```go

// UpdateInvoices is a function to update a single record from invoices table in the main database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateInvoices(ctx context.Context,   argInvoiceID int32,        updated *model.Invoices) (result *model.Invoices, RowsAffected int64, err error) {

   result = &model.Invoices{}
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

// DeleteInvoices is a function to delete a single record from invoices table in the main database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteInvoices(ctx context.Context,  argInvoiceID int32,        ) (rowsAffected int64, err error) {

    record := &model.Invoices{}
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
