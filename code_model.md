## Go Struct Generation
`gen` will generate a `.go` file per table, mapping every column to a field in struct. The package name can be set with the    
`--model` flag. The model code generation can be customized with the following flags. 


### Struct Tag Customizations
- `--gorm` - Add GORM struct tag to fields  
- `--json` - Add json struct tag to fields  
- `--no-json` - Disable json struct tag
- `--json-fmt=` - json name format [snake | camel | lower_camel | none]
- `--protobuf` - Add protobuf struct tag to fields
- `--db` - Add db struct tag to fields

### Handling Null Columns
When columns within a db table are nullable, `gen` can generate code with nullable types using the `--guregu` flag, or it can use the default go type.
For more information on the types used check out [guregu null](https://github.com/guregu/null). 

    
```go
/*
DB Table Details
-------------------------------------


CREATE TABLE "albums"
(
    [AlbumId] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    [Title] NVARCHAR(160)  NOT NULL,
    [ArtistId] INTEGER  NOT NULL,
    FOREIGN KEY ([ArtistId]) REFERENCES "artists" ([ArtistId])
		ON DELETE NO ACTION ON UPDATE NO ACTION
)

JSON Sample
-------------------------------------
{    "title": "FBQslWHyXcrTdtrEZpqMgCklk",    "artist_id": 8,    "album_id": 8}
*/
```

```go
// Album struct is a row record of the albums table in the main database
type Album struct {
	//[ 0] AlbumId                                        integer              null: false  primary: true   auto: true   col: integer         len: -1      default: []
	AlbumID int `gorm:"primary_key;AUTO_INCREMENT;column:AlbumId;type:INTEGER;" json:"album_id" db:"AlbumId" protobuf:"int32,0,opt,name=album_id"`
	//[ 1] Title                                          nvarchar(160)        null: false  primary: false  auto: false  col: nvarchar        len: 160     default: []
	Title string `gorm:"column:Title;type:NVARCHAR(160);size:160;" json:"title" db:"Title" protobuf:"string,1,opt,name=title"`
	//[ 2] ArtistId                                       integer              null: false  primary: false  auto: false  col: integer         len: -1      default: []
	ArtistID int `gorm:"column:ArtistId;type:INTEGER;" json:"artist_id" db:"ArtistId" protobuf:"int32,2,opt,name=artist_id"`
}
```

### Other Functions
Extra funcs are generated for each struct. The funcs are intended for data validation.  

```go
// TableName sets the insert table name for this struct type
func (a *Album) TableName() string {
	return "albums"
}

// BeforeSave invoked before saving, return an error if field is not populated.
func (a *Album) BeforeSave() error {
	return nil
}

// Prepare invoked before saving, can be used to populate fields etc.
func (a *Album) Prepare() {
}

// Validate invoked before performing action, return an error if field is not populated.
func (a *Album) Validate(action Action) error {
	return nil
}
```
