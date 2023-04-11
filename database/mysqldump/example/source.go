package main

import (
	"os"

	"github.com/xishengcai/ganni-tool/database/mysqldump"
)

func main() {

	dns := "root:123456@tcp(localhost:3306)/sss?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	f, err := os.Open("dump.sql")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()
	_ = mysqldump.Source(
		dns,
		f,
		// mysqldump.WithDeleteDB(), // Option: Delete db before create (Default: Not delete db)
	)
}
