package parser

import (
	"errors"
	"strings"
)

var ErrInvalidSQL = errors.New("invalid sql")

// Parse 解析简化 SQL。
func Parse(sql string) (Statement, error) {
	tokens := Tokenize(sql)
	if len(tokens) == 0 {
		return nil, ErrInvalidSQL
	}

	switch strings.ToUpper(tokens[0]) {
	case "CREATE":
		if len(tokens) >= 3 && strings.ToUpper(tokens[1]) == "TABLE" {
			return &CreateTableStmt{Table: tokens[2]}, nil
		}
	case "INSERT":
		// INSERT INTO t key value
		if len(tokens) >= 5 && strings.ToUpper(tokens[1]) == "INTO" {
			return &InsertStmt{Table: tokens[2], Key: tokens[3], Value: tokens[4]}, nil
		}
	case "SELECT":
		// SELECT * FROM t WHERE key = k
		if len(tokens) >= 8 && strings.ToUpper(tokens[1]) == "*" && strings.ToUpper(tokens[2]) == "FROM" && strings.ToUpper(tokens[4]) == "WHERE" && strings.EqualFold(tokens[5], "key") {
			return &SelectStmt{Table: tokens[3], Key: tokens[7]}, nil
		}
		// SELECT RANGE FROM t start end
		if len(tokens) >= 6 && strings.ToUpper(tokens[1]) == "RANGE" && strings.ToUpper(tokens[2]) == "FROM" {
			return &SelectRangeStmt{Table: tokens[3], Start: tokens[4], End: tokens[5]}, nil
		}
	case "BEGIN", "COMMIT", "ROLLBACK":
		return &TxStmt{Action: strings.ToUpper(tokens[0])}, nil
	}

	return nil, ErrInvalidSQL
}
