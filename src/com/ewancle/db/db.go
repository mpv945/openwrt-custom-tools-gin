package db

import (
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// InitDB driverName: [pgx, mysql]
func InitDB(host string, port int, user string, password string, dbname string, driverName string) {

	/*
			| 转换               | 方法                   |
			| ---------------- | ---------------------- |
			| int → int64      | `int64(a)`             |
			| int → string     | `strconv.Itoa()`       |
			| string → int     | `strconv.Atoi()`       |
			| string → int64   | `strconv.ParseInt()`   |
			| int64 → string   | `strconv.FormatInt()`  |
			| string → float   | `strconv.ParseFloat()` |
			| []byte → string  | `string(b)`            |
			| string → []byte  | `[]byte(s)`            |
			| interface → type | `v.(type)`             |
		interface 类型转换（非常重要）
			var i interface{} = "hello"

			s, ok := i.(string)

			if ok {
				fmt.Println(s)
			}
		时间类型转换
			string → time
				import "time"
				t, err := time.Parse("2006-01-02", "2025-01-01")
			time → string
				s := t.Format("2006-01-02")
	*/
	dsn := user + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + dbname + "?charset=utf8mb4&parseTime=true&loc=Local"
	//dsn := "postgres://postgres:password@localhost:5432/demo?sslmode=disable"

	var err error
	//DB, err = sqlx.Connect("mysql", dsn)
	//DB, err = sqlx.Connect("pgx", dsn)
	DB, err = sqlx.Connect(driverName, dsn)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}

	// 生产环境必须配置连接池
	DB.SetMaxOpenConns(50)
	DB.SetMaxIdleConns(20)
	DB.SetConnMaxLifetime(time.Hour)
	//DB.SetConnMaxLifetime(time.Minute * 3)

	log.Println("MySQL connected")
}
