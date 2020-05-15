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
// GetAllAlbums is a function to get a slice of record(s) from albums table in the main database
// @Summary Get list of Album
// @Tags Album
// @Description GetAllAlbum is a handler to get a slice of record(s) from albums table in the main database
// @Accept  json
// @Produce  json
// @Param   page     query    int     false        "page requested (defaults to 0)"
// @Param   pagesize query    int     false        "number of records in a page  (defaults to 20)"
// @Param   order    query    string  false        "db sort order column"
// @Success 200 {object} apis.PagedResults{data=[]models.Album}
// @Failure 400 {object} apis.HTTPError
// @Failure 404 {object} apis.HTTPError
// @Router /albums [get]
// http http://localhost:8080/albums?page=0&pagesize=20
func GetAllAlbums(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	pagesize, err := readInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	order := r.FormValue("order")

	records, totalRows, err := daos.GetAllAlbums(r.Context(), page, pagesize, order)
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
// GetAlbum is a function to get a single record to albums table in the main database
// @Summary Get record from table Album by  argAlbumID
// @Tags Album
// @ID argAlbumID

// @Description GetAlbum is a function to get a single record to albums table in the main database
// @Accept  json
// @Produce  json
// @Param  argAlbumID path int true "AlbumId"
// @Success 200 {object} models.Album
// @Failure 400 {object} apis.HTTPError
// @Failure 404 {object} apis.HTTPError "ErrNotFound, db record for id not found - returns NotFound HTTP 404 not found error"
// @Router /albums/{argAlbumID} [get]
// http http://localhost:8080/albums/1
func GetAlbum(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	argAlbumID, err := parseInt(ps, "argAlbumID")
	if err != nil {
		returnError(w, r, err)
		return
	}

	record, err := daos.GetAlbum(r.Context(), argAlbumID)
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeJSON(w, record)
}
```

## Create record
```go

// AddAlbum add to add a single record to albums table in the main database
// @Summary Add an record to albums table
// @Description add to add a single record to albums table in the main database
// @Tags Album
// @Accept  json
// @Produce  json
// @Param Album body models.Album true "Add Album"
// @Success 200 {object} models.Album
// @Failure 400 {object} apis.HTTPError
// @Failure 404 {object} apis.HTTPError
// @Router /albums [post]
// echo '{"title": "FBQslWHyXcrTdtrEZpqMgCklk","artist_id": 8,"album_id": 8}' | http POST http://localhost:8080/albums
func AddAlbum(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	album := &models.Album{}

	if err := readJSON(r, album); err != nil {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	if err := album.BeforeSave(); err != nil {
		returnError(w, r, daos.ErrBadParams)
	}

	album.Prepare()

	if err := album.Validate(models.Create); err != nil {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	var err error
	album, _, err = daos.AddAlbum(r.Context(), album)
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeJSON(w, album)
}
```

## Update record
```go

// UpdateAlbum Update a single record from albums table in the main database
// @Summary Update an record in table albums
// @Description Update a single record from albums table in the main database
// @Tags Album
// @Accept  json
// @Produce  json
// @Param  argAlbumID path int true "AlbumId"
// @Param  Album body models.Album true "Update Album record"
// @Success 200 {object} models.Album
// @Failure 400 {object} apis.HTTPError
// @Failure 404 {object} apis.HTTPError
// @Router /albums/{argAlbumID} [patch]
// echo '{"title": "FBQslWHyXcrTdtrEZpqMgCklk","artist_id": 8,"album_id": 8}' | http PATCH http://localhost:8080/albums/1
func UpdateAlbum(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	argAlbumID, err := parseInt(ps, "argAlbumID")
	if err != nil {
		returnError(w, r, err)
		return
	}

	album := &models.Album{}
	if err := readJSON(r, album); err != nil {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	if err := album.BeforeSave(); err != nil {
		returnError(w, r, daos.ErrBadParams)
	}

	album.Prepare()

	if err := album.Validate(models.Update); err != nil {
		returnError(w, r, daos.ErrBadParams)
		return
	}

	album, _, err = daos.UpdateAlbum(r.Context(),
		argAlbumID,
		album)
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeJSON(w, album)
}
```

## Delete record
```go

// DeleteAlbum Delete a single record from albums table in the main database
// @Summary Delete a record from albums
// @Description Delete a single record from albums table in the main database
// @Tags Album
// @Accept  json
// @Produce  json
// @Param  argAlbumID path int true "AlbumId"
// @Success 204 {object} models.Album
// @Failure 400 {object} apis.HTTPError
// @Failure 500 {object} apis.HTTPError
// @Router /albums/{argAlbumID} [delete]
// http DELETE http://localhost:8080/albums/1
func DeleteAlbum(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	argAlbumID, err := parseInt(ps, "argAlbumID")
	if err != nil {
		returnError(w, r, err)
		return
	}

	rowsAffected, err := daos.DeleteAlbum(r.Context(), argAlbumID)
	if err != nil {
		returnError(w, r, err)
		return
	}

	writeRowsAffected(w, rowsAffected)
}

```
