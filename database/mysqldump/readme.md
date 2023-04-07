# mysql dump
this project is forked from "https://github.com/jarvanstack/mysqldump"

## example
- dump
```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xishengcai/ganni-tool/database/mysqldump"
)

func main() {

	// contain database
	//dns := "root:rootpasswd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	// ignore database name
	dns := "root:123456@tcp(localhost:3306)/?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"

	f, _ := os.Create(fmt.Sprintf("backup-20060102T150405.sql"))
	defer func() {
		_ = f.Close()
	}()
	_ = mysqldump.Dump(
		dns,                       // DNS
		mysqldump.WithDropTable(), // Option: Delete table before create (Default: Not delete table)
		mysqldump.WithData(),      // Option: Dump Data (Default: Only dump table schema)
		//mysqldump.WithTables("test"), // Option: Dump Tables (Default: All tables)
		mysqldump.WithWriter(f), // Option: Writer (Default: os.Stdout)
		//mysqldump.WithDBs("dc3"),     // Option: Dump Dbs (Default: db in dns)
		mysqldump.WithAllDatabases(),
	)
}

```

- source
```go
package main

import (
	"os"

	"github.com/xishengcai/ganni-tool/database/mysqldump"
)

func main() {

	dns := "root:rootpasswd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	f, _ := os.Open("dump.sql")
	defer func() {
		_ = f.Close()
	}()
	_ = mysqldump.Source(
		dns,
		f,
		// mysqldump.WithDeleteDB(), // Option: Delete db before create (Default: Not delete db)
	)
}

```