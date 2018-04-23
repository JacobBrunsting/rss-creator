package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

type SqlCondition struct {
	Column     string
	Comparator string
	Arg        interface{}
}

// Takes an array of SQL conditions, and returns a SQL WHERE statement with
// an array of arguments. Excludes SQL conditions where Arg is the zero value
func SqlWhere(conditions []SqlCondition) (string, []interface{}) {
	formatted := []string{}
	args := make([]interface{}, 0)

	for _, condition := range conditions {
		col := condition.Column
		comp := condition.Comparator
		arg := condition.Arg

		strVal := ""

		switch a := arg.(type) {
		case string:
			strVal = a
		case int:
			if a == 0 {
				break
			}
			strVal = strconv.Itoa(a)
		case time.Time:
			if a.IsZero() {
				break
			}
			strVal = a.Format(TimeFormat)
		default:
			log.Printf("type %T not supported by ConstructSqlWhere", a)
		}

		if strVal != "" {
			formatted = append(formatted, fmt.Sprintf("%v %v ?", col, comp))
			args = append(args, strVal)
		}
	}

	return fmt.Sprintf("WHERE %v", strings.Join(formatted, " AND ")), args
}
