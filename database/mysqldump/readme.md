# mysql dump
this project is forked from "https://github.com/jarvanstack/mysqldump"

## example
- dump
```go
package main

import (
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


## 内置数据库介绍
- mysql
mysql：这个是mysql的核心数据库，类似于sql server中的master表，主要负责存储数据库的用户、权限设置、关键字等mysql
自己需要使用的控制和管理信息。不可以删除，如果对mysql不是很了解，也不要轻易修改这个数据库里面的表信息。

## 路线图
- 备份镜像制作
使用mysql client 镜像 + shell 脚本

- 备份镜像charts 制作
使用corn job 镜像进行定时备份

## issue:
- 1. 如何创建database
- 2. 如何增量备份
- 3. 如何备份视图