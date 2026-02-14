package parser

// Statement SQL 语句抽象。
type Statement interface {
	statementName() string
}

// CreateTableStmt CREATE TABLE。
type CreateTableStmt struct {
	Table string
}

func (s *CreateTableStmt) statementName() string { return "create_table" }

// InsertStmt INSERT。
type InsertStmt struct {
	Table string
	Key   string
	Value string
}

func (s *InsertStmt) statementName() string { return "insert" }

// SelectStmt SELECT。
type SelectStmt struct {
	Table string
	Key   string
}

func (s *SelectStmt) statementName() string { return "select" }

// SelectRangeStmt 范围查询。
type SelectRangeStmt struct {
	Table string
	Start string
	End   string
}

func (s *SelectRangeStmt) statementName() string { return "select_range" }

// TxStmt 事务语句。
type TxStmt struct {
	Action string
}

func (s *TxStmt) statementName() string { return "tx" }
