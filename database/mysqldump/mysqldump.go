package mysqldump

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 打印 日志 行数
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

type dumpOption struct {
	// 导出表数据
	isData bool
	// 导出指定数据库, 与 WithAllDatabases 互斥, WithAllDatabases 优先级高
	dbs []string
	// 导出全部数据库
	isAllDB bool
	// 导出指定表, 与 isAllTables 互斥, isAllTables 优先级高
	tables []string
	// 导出全部表
	isAllTable bool
	// 是否删除表
	isDropTable bool

	// writer 默认为 os.Stdout
	writer io.Writer
}

type DumpOption func(*dumpOption)

// WithDropTable 删除表
func WithDropTable() DumpOption {
	return func(option *dumpOption) {
		option.isDropTable = true
	}
}

// WithData 导出表数据
func WithData() DumpOption {
	return func(option *dumpOption) {
		option.isData = true
	}
}

// WithAllDatabases 导出全部数据库
func WithAllDatabases() DumpOption {
	return func(option *dumpOption) {
		option.isAllDB = true
	}
}

// WithDBs 导出指定数据库, 与 WithAllDatabases 互斥, WithAllDatabases 优先级高
func WithDBs(databases ...string) DumpOption {
	return func(option *dumpOption) {
		option.dbs = databases
	}
}

// WithTables 导出指定表, 与 WithAllTables 互斥, WithAllTables 优先级高
func WithTables(tables ...string) DumpOption {
	return func(option *dumpOption) {
		option.tables = tables
	}
}

// WithAllTable 导出全部表
func WithAllTable() DumpOption {
	return func(option *dumpOption) {
		option.isAllTable = true
	}
}

// WithWriter 导出到指定 writer
func WithWriter(writer io.Writer) DumpOption {
	return func(option *dumpOption) {
		option.writer = writer
	}
}

// Dump 导出SQL
func Dump(dns string, opts ...DumpOption) error {
	var err error

	var o dumpOption

	for _, opt := range opts {
		opt(&o)
	}

	if len(o.dbs) == 0 {
		// 默认包含dns中的数据库
		dbName, err := GetDBNameFromDNS(dns)
		if err != nil {
			log.Printf("[error] %v \n", err)
			return err
		}
		o.dbs = []string{
			dbName,
		}
	}

	if len(o.tables) == 0 {
		// 默认包含全部表
		o.isAllTable = true
	}

	if o.writer == nil {
		// 默认输出到 os.Stdout
		o.writer = os.Stdout
	}

	buf := bufio.NewWriter(o.writer)
	defer buf.Flush()

	// 打印 Header
	start := time.Now()
	buf.WriteString("-- ----------------------------\n")
	buf.WriteString("-- MySQL Database Dump\n")
	buf.WriteString("-- Start Time: " + start.Format("2006-01-02 15:04:05") + "\n")
	buf.WriteString("-- ----------------------------\n")
	buf.WriteString("\n\n")

	// 连接数据库
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Printf("[error] %v \n", err)
		return err
	}
	defer db.Close()

	// 1. 获取数据库
	var dbs []string
	if o.isAllDB {
		dbs, err = getDBs(db)
		if err != nil {
			log.Printf("[error] %v \n", err)
			return err
		}
	} else {
		dbs = o.dbs
	}

	// 2. 获取表
	for _, dbStr := range dbs {
		buf.WriteString("--\n")
		buf.WriteString("-- Current Database: `" + dbStr + "`\n")
		buf.WriteString("--\n\n")
		buf.WriteString(fmt.Sprintf("CREATE DATABASE /*!32312 IF NOT EXISTS*/ `%s` /*!40100 DEFAULT "+
			"CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;\n", dbStr))
		buf.WriteString("USE `" + dbStr + "`;\n")

		_, err = db.Exec("USE `" + dbStr + "`")
		if err != nil {
			log.Printf("[error] %v \n", err)
			return err
		}

		var tables []string
		if o.isAllTable {
			tmp, err := getAllTables(db)
			if err != nil {
				log.Printf("[error] %v \n", err)
				return err
			}
			tables = tmp
		} else {
			tables = o.tables
		}

		// 3. 导出表
		log.Printf("db: %s tables: %s \n", dbStr, tables)
		for _, table := range tables {
			// 删除表
			if o.isDropTable {
				buf.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table))
			}

			// 导出表结构
			err = writeTableStruct(db, table, buf)
			if err != nil {
				return err
			}

			// 导出表数据
			if o.isData {
				err = writeTableData(db, table, buf)
				if err != nil {
					log.Printf("[error] %v \n", err)
					return err
				}
			}
		}

	}

	// 导出每个表的结构和数据

	buf.WriteString("-- ----------------------------\n")
	buf.WriteString("-- Dumped by mysqldump2\n")
	buf.WriteString("-- Cost Time: " + time.Since(start).String() + "\n")
	buf.WriteString("-- ----------------------------\n")

	return nil
}

func getCreateTableSQL(db *sql.DB, table string) (string, error) {
	var createTableSQL string
	err := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`;", table)).Scan(&table, &createTableSQL)
	if err != nil {
		log.Println("[error]: ", err)
		return "", err
	}
	// IF NOT EXISTS
	createTableSQL = strings.Replace(createTableSQL, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS", 1)
	return createTableSQL, nil
}

func getDBs(db *sql.DB) ([]string, error) {
	var dbs []string
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var db string
		err = rows.Scan(&db)
		if err != nil {
			return nil, err
		}
		if !ISInternalDatabase(db) {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}

func getAllTables(db *sql.DB) ([]string, error) {
	var tables []string
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func writeTableStruct(db *sql.DB, table string, buf *bufio.Writer) error {
	// 导出表结构
	buf.WriteString("-- ----------------------------\n")
	buf.WriteString(fmt.Sprintf("-- Table structure for %s\n", table))
	buf.WriteString("-- ----------------------------\n")

	createTableSQL, err := getCreateTableSQL(db, table)
	if err != nil {
		return err
	}
	buf.WriteString(createTableSQL)
	buf.WriteString(";")

	buf.WriteString("\n\n")
	buf.WriteString("\n\n")
	return nil
}

func writeTableData(db *sql.DB, table string, buf *bufio.Writer) error {

	// 导出表数据
	buf.WriteString("-- ----------------------------\n")
	buf.WriteString(fmt.Sprintf("-- Records of %s\n", table))
	buf.WriteString("-- ----------------------------\n")

	lineRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		log.Printf("[error] %v \n", err)
		return err
	}
	defer lineRows.Close()

	var columns []string
	columns, err = lineRows.Columns()
	if err != nil {
		log.Printf("[error] %v \n", err)
		return err
	}
	columnTypes, err := lineRows.ColumnTypes()
	if err != nil {
		log.Printf("[error] %v \n", err)
		return err
	}

	var values [][]interface{}
	for lineRows.Next() {
		row := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range columns {
			rowPointers[i] = &row[i]
		}
		err = lineRows.Scan(rowPointers...)
		if err != nil {
			log.Printf("[error] %v \n", err)
			return err
		}
		values = append(values, row)
	}

	for _, row := range values {
		ssql := "INSERT INTO `" + table + "` VALUES ("

		for i, col := range row {
			if col == nil {
				ssql += "NULL"
			} else {
				Type := columnTypes[i].DatabaseTypeName()
				// 去除 UNSIGNED 和空格
				Type = strings.Replace(Type, "UNSIGNED", "", -1)
				Type = strings.Replace(Type, " ", "", -1)
				switch Type {
				case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT":
					if bs, ok := col.([]byte); ok {
						ssql += fmt.Sprintf("%s", string(bs))
					} else {
						ssql += fmt.Sprintf("%d", col)
					}
				case "FLOAT", "DOUBLE":
					if bs, ok := col.([]byte); ok {
						ssql += fmt.Sprintf("%s", string(bs))
					} else {
						ssql += fmt.Sprintf("%f", col)
					}
				case "DECIMAL", "DEC":
					ssql += fmt.Sprintf("%s", col)

				case "DATE":
					t, ok := col.(time.Time)
					if !ok {
						log.Println("DATE 类型转换错误")
						return err
					}
					ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02"))
				case "DATETIME":
					t, ok := col.(time.Time)
					if !ok {
						log.Println("DATETIME 类型转换错误")
						return err
					}
					ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
				case "TIMESTAMP":
					t, ok := col.(time.Time)
					if !ok {
						log.Println("TIMESTAMP 类型转换错误")
						return err
					}
					ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
				case "TIME":
					t, ok := col.([]byte)
					if !ok {
						log.Println("TIME 类型转换错误")
						return err
					}
					ssql += fmt.Sprintf("'%s'", string(t))
				case "YEAR":
					t, ok := col.([]byte)
					if !ok {
						log.Println("YEAR 类型转换错误")
						return err
					}
					ssql += fmt.Sprintf("%s", string(t))
				case "CHAR", "VARCHAR", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT":
					ssql += fmt.Sprintf("'%s'", strings.Replace(fmt.Sprintf("%s", col), "'", "''", -1))
				case "BIT", "BINARY", "VARBINARY", "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB":
					ssql += fmt.Sprintf("0x%X", col)
				case "ENUM", "SET":
					ssql += fmt.Sprintf("'%s'", col)
				case "BOOL", "BOOLEAN":
					if col.(bool) {
						ssql += "true"
					} else {
						ssql += "false"
					}
				case "JSON":
					ssql += fmt.Sprintf("'%s'", col)
				default:
					// unsupported type
					log.Printf("unsupported type: %s", Type)
					return fmt.Errorf("unsupported type: %s", Type)
				}
			}
			if i < len(row)-1 {
				ssql += ","
			}
		}
		ssql += ");\n"
		buf.WriteString(ssql)
	}

	buf.WriteString("\n\n")
	return nil
}
