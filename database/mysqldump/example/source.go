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
