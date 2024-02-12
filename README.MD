# Phoenix
## Description
Phoenix is a database migration tool to keep track of database changes over time.  
## Supported Databases
- MySQL
- PostgreSQL

## Installation
```bash
go get github.com/bernhardkrug/phoenix
```

## Usage
You can trigger your database migrations directly after connecting to your database.  
All you need to provide is the database interface and the database type. 
Optional you can configure Phoenix by using the provided ConfigFunctiuns. See chapter Configuration for more details.

### Example:

```go
package main

import (
	"database/sql"
	"github.com/bernhardkrug/phoenix"
)

func main() {
	var db *sql.DB
	db = ... // connect to your database
	
	phoenix.Rise(db, phoenix.Postgres)
}
```
### Default Configuration
By starting the migration Phoenix looks for migration files in the directory ``'sql'`` inside of the project's root folder.
The table where Phoenix stores the migration history is called 'phoenix_history'. 
It also does not set a schema per default.

If one of these defaults does not fit your needs you can configure Phoenix by using the provided ConfigFunctions.
## Configuration
You can use one of the following ConfigFunctions to configure Phoenix:  
``WithImportFolder(folder string)``: Overides the folder where Phoenix looks for migration scripts.  
``WithSchema(schemaName string)``: If you are using a schema you can specify it here.  
``WithTableName(tableName string)``: Specifies the table name, where Phoenix stores the migration history.

