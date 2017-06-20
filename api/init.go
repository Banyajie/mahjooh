package api

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"chess_alg_jx/logger"
)

const limitClause = " limit ?,?"

var (
	db      *sqlx.DB
	countRe *regexp.Regexp
)

func init() {
	countRe = regexp.MustCompile(`(?si:select(.*?(?s))\W+from(\W+.*(?s)))`)
}

func genSqlCount(oldsql string) string {
	return countRe.ReplaceAllString(oldsql, "select count(*) from $2")
}

func getTotal(sql, cond string, args []interface{}) int {
	res := 0
	sqlStr := genSqlCount(sql) + cond
	logger.Info("regex sql:", sqlStr)
	if err := db.QueryRow(sqlStr, args...).Scan(&res); err != nil {
		logger.Error("total count failed:", err)
	}
	return res
}

func initDB(connStr string, db **sqlx.DB) error {
	var err error
	*db, err = sqlx.Open("mysql", connStr)
	if err != nil {
		return err
	}
	err = (*db).Ping()
	if err != nil {
		return err
	}
	(*db).SetMaxIdleConns(100)
	(*db).SetConnMaxLifetime(time.Minute)
	return nil
}

func InitDB(constr string) error {
	return initDB(constr, &db)
}

func isZero(v reflect.Value) bool {
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

func genWhereCond(val interface{}) (string, []interface{}) {
	var (
		arrSF []reflect.StructField
		arrV  []reflect.Value
		args  = make([]interface{}, 0)
	)

	value := reflect.ValueOf(val)

	switch value.Kind() {
	case reflect.Ptr:
		value = reflect.Indirect(value)
		fallthrough
	case reflect.Struct:
		t := value.Type()
		l := value.NumField()
		for i := 0; i < l; i++ {
			//skip zero value field
			if isZero(value.Field(i)) {
				continue
			}
			//skip field has no tag
			if len(t.Field(i).Tag) == 0 {
				continue
			}
			arrSF = append(arrSF, t.Field(i))
			arrV = append(arrV, value.Field(i))
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", value.Kind()))
	}

	sqlStr := ""
	if len(arrV) > 0 {
		sqlStr += " where "
	}
	for i, e := range arrSF {
		if i > 0 {
			sqlStr += " and "
		}
		sqlStr += string(e.Tag)
		if e.Type.String() == "time.Time" && e.Name == "From" {
			sqlStr += " >= ? "
		} else if e.Type.String() == "time.Time" && e.Name == "To" {
			sqlStr += " <= ? "
		} else {
			sqlStr += " = ? "
		}
	}
	for _, e := range arrV {
		args = append(args, e.Interface())
	}

	return sqlStr, args
}
