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
# GORM
```go
// GetAllAlbums is a function to get a slice of record(s) from albums table in the main database
// params - page     - page requested (defaults to 0)
// params - pagesize - number of records in a page  (defaults to 20)
// params - order    - db sort order column
// error - ErrNotFound, db Find error
func GetAllAlbums(ctx context.Context, page, pagesize int64, order string) (albums []*models.Album, totalRows int, err error) {

	albums = []*models.Album{}

	albumsOrm := DB.Model(&models.Album{})
	albumsOrm.Count(&totalRows)

	if page > 0 {
		offset := (page - 1) * pagesize
		albumsOrm = albumsOrm.Offset(offset).Limit(pagesize)
	} else {
		albumsOrm = albumsOrm.Limit(pagesize)
	}

	if order != "" {
		albumsOrm = albumsOrm.Order(order)
	}

	if err = albumsOrm.Find(&albums).Error; err != nil {
		err = ErrNotFound
		return nil, -1, err
	}

	return albums, totalRows, nil
}
```

# sqlx
```go
func GetAllAlbums(ctx context.Context, page, pagesize int64, order string) (albums []*models.Album, totalRows int, err error) {
	sql := "SELECT * FROM albums"

	if order != "" {
		sql = sql + " order by " + order
	}
	sql = fmt.Sprintf("%s LIMIT %d, %d", sql, page, pagesize)
	fmt.Printf("%s\n", sql)
	err = DB.SelectContext(ctx, &albums, sql, page, pagesize)
	return albums, len(albums), err
}

```

## Retrieve record
# GORM
```go
// GetAlbum is a function to get a single record to albums table in the main database
// error - ErrNotFound, db Find error
func GetAlbum(ctx context.Context, argAlbumID int) (record models.Album, err error) {
	if err = DB.First(&record, argAlbumID).Error; err != nil {
		err = ErrNotFound
		return record, err
	}

	return record, nil
}
```

# SQLX
```go
func GetAlbum(ctx context.Context, argAlbumID int) (record *models.Album, err error) {
	sql := "SELECT * FROM albums WHERE AlbumId=?"
	record = &models.Album{}
	err = DB.GetContext(ctx, record, sql, argAlbumID)
	if err != nil {
		return nil, err
	}
	return record, nil
}
```

## Create record
```go
// AddAlbum is a function to add a single record to albums table in the main database
// error - ErrInsertFailed, db save call failed
func AddAlbum(ctx context.Context, record *models.Album) (result *models.Album, RowsAffected int64, err error) {
	db := DB.Save(record)
	if err = db.Error; err != nil {
		return nil, -1, ErrInsertFailed
	}

	return record, db.RowsAffected, nil
}
```
# SQLX
```go
func AddAlbum(ctx context.Context, record *models.Album) (result *models.Album, RowsAffected int64, err error) {
	sql := "INSERT INTO albums ( Title,  ArtistId) values ( $1, $2 )"
	dbResult := DB.MustExecContext(ctx, sql, record.Title, record.ArtistID)
	id, err := dbResult.LastInsertId()
	rows, err := dbResult.RowsAffected()
	record.AlbumID = int(id)
	return record, rows, err
```


## Update record
```go

// UpdateAlbum is a function to update a single record from albums table in the main database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateAlbum(ctx context.Context, argAlbumID int, updated *models.Album) (result *models.Album, RowsAffected int64, err error) {

	result = &models.Album{}
	db := DB.First(result, argAlbumID)
	if err = db.Error; err != nil {
		return nil, -1, ErrNotFound
	}

	if err = dbmeta.Copy(result, updated); err != nil {
		return nil, -1, ErrUpdateFailed
	}

	db = db.Save(result)
	if err = db.Error; err != nil {
		return nil, -1, ErrUpdateFailed
	}

	return result, db.RowsAffected, nil
}
```
# SQLX
```go
func UpdateAlbum(ctx context.Context, argAlbumID int, updated *models.Album) (result *models.Album, RowsAffected int64, err error) {
	sql := "UPDATE albums set Title = $1, ArtistId = $2 WHERE AlbumId = $3"
	dbResult := DB.MustExecContext(ctx, sql, updated.Title, updated.ArtistID, argAlbumID)
	id, err := dbResult.LastInsertId()
	rows, err := dbResult.RowsAffected()
	updated.AlbumID = argAlbumID
	return updated, rows, err
}
```


## Delete record
```go

// DeleteAlbum is a function to delete a single record from albums table in the main database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteAlbum(ctx context.Context, argAlbumID int) (rowsAffected int64, err error) {

	record := &models.Album{}
	db := DB.First(record, argAlbumID)
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
# SQLX
```go
func DeleteAlbum(ctx context.Context, argAlbumID int) (rowsAffected int64, err error) {
	sql := "DELETE FROM albums where AlbumId = $1"
	result := DB.MustExecContext(ctx, sql, argAlbumID)
	return result.RowsAffected()
}
```
