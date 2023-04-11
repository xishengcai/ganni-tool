package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xishengcai/ganni-tool/database/mysqldump"
)

var (
	host     string
	port     int
	user     string
	password string
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "")
	flag.IntVar(&port, "port", 3306, "")
	flag.StringVar(&user, "user", "root", "")
	flag.StringVar(&password, "pwd", "123456", "")
}

func main() {

	// contain database
	//dns := "root:rootpasswd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	// ignore database name
	dns := "root:123456@tcp(localhost:3306)/?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"

	//f, err := os.Create(fmt.Sprintf("backup-20060102T150405.sql"))
	f, err := os.Create(fmt.Sprintf("dump.sql"))
	if err != nil {
		panic(err)
	}
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
