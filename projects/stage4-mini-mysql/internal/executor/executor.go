package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/txn"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

var (
	ErrTableNotFound = errors.New("table not found")
)

// Engine 执行引擎。
type Engine struct {
	mu       sync.RWMutex
	tables   map[string]*storage.BPlusTree
	txMgr    *txn.Manager
	activeTx map[string]*txn.Transaction
}

// NewEngine 创建执行引擎。
func NewEngine(txMgr *txn.Manager) *Engine {
	return &Engine{tables: make(map[string]*storage.BPlusTree), txMgr: txMgr, activeTx: make(map[string]*txn.Transaction)}
}

// CreateTable 创建表。
func (e *Engine) CreateTable(name string, tree *storage.BPlusTree) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tables[name] = tree
}

// Execute 执行 SQL 语句。
func (e *Engine) Execute(ctx context.Context, sessionID string, statement parser.Statement) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	switch s := statement.(type) {
	case *parser.CreateTableStmt:
		return e.execCreateTable(s)
	case *parser.InsertStmt:
		return e.execInsert(ctx, sessionID, s)
	case *parser.SelectStmt:
		return e.execSelect(ctx, s)
	case *parser.SelectRangeStmt:
		return e.execSelectRange(ctx, s)
	case *parser.TxStmt:
		return e.execTx(ctx, sessionID, s)
	default:
		return "", parser.ErrInvalidSQL
	}
}

func (e *Engine) execCreateTable(stmt *parser.CreateTableStmt) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.tables[stmt.Table]; ok {
		return "OK", nil
	}
	e.tables[stmt.Table] = storage.NewBPlusTree(16, nil)
	return "OK", nil
}

func (e *Engine) execInsert(ctx context.Context, sessionID string, stmt *parser.InsertStmt) (string, error) {
	tree, err := e.getTable(stmt.Table)
	if err != nil {
		return "", err
	}

	e.mu.Lock()
	tx := e.activeTx[sessionID]
	e.mu.Unlock()
	if tx != nil {
		if err = e.txMgr.Put(ctx, tx, []byte(stmt.Key), []byte(stmt.Value)); err != nil {
			return "", err
		}
		return "OK", nil
	}
	if err = tree.Insert(ctx, []byte(stmt.Key), []byte(stmt.Value)); err != nil {
		return "", err
	}
	return "OK", nil
}

func (e *Engine) execSelect(ctx context.Context, stmt *parser.SelectStmt) (string, error) {
	tree, err := e.getTable(stmt.Table)
	if err != nil {
		return "", err
	}
	value, err := tree.Search(ctx, []byte(stmt.Key))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return "NULL", nil
		}
		return "", err
	}
	return string(value), nil
}

func (e *Engine) execSelectRange(ctx context.Context, stmt *parser.SelectRangeStmt) (string, error) {
	tree, err := e.getTable(stmt.Table)
	if err != nil {
		return "", err
	}
	iterator, err := tree.RangeScan(ctx, []byte(stmt.Start), []byte(stmt.End))
	if err != nil {
		return "", err
	}
	result := ""
	for iterator.Next() {
		item := iterator.Item()
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("%s=%s", string(item.Key), string(item.Value))
	}
	if result == "" {
		return "NULL", nil
	}
	return result, nil
}

func (e *Engine) execTx(ctx context.Context, sessionID string, stmt *parser.TxStmt) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch stmt.Action {
	case "BEGIN":
		tx, err := e.txMgr.Begin(ctx)
		if err != nil {
			return "", err
		}
		e.activeTx[sessionID] = tx
		return "OK", nil
	case "COMMIT":
		tx := e.activeTx[sessionID]
		if tx == nil {
			return "NO_TX", nil
		}
		if err := e.txMgr.Commit(ctx, tx); err != nil {
			return "", err
		}
		delete(e.activeTx, sessionID)
		return "OK", nil
	case "ROLLBACK":
		tx := e.activeTx[sessionID]
		if tx == nil {
			return "NO_TX", nil
		}
		if err := e.txMgr.Abort(ctx, tx); err != nil {
			return "", err
		}
		delete(e.activeTx, sessionID)
		return "OK", nil
	default:
		return "", parser.ErrInvalidSQL
	}
}

func (e *Engine) getTable(name string) (*storage.BPlusTree, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	table, ok := e.tables[name]
	if !ok {
		return nil, ErrTableNotFound
	}
	return table, nil
}

// BootstrapDefaultTable 初始化默认表。
func BootstrapDefaultTable(engine *Engine, tree *storage.BPlusTree) {
	engine.CreateTable("kv", tree)
}

// NewWithDefaults 创建默认引擎。
func NewWithDefaults(tree *storage.BPlusTree, walWriter *wal.Writer) *Engine {
	txMgr := txn.NewManager(walWriter, tree)
	engine := NewEngine(txMgr)
	BootstrapDefaultTable(engine, tree)
	return engine
}
