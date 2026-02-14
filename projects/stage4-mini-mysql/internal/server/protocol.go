package server

import "strings"

// normalizeSQL 清理 SQL 输入。
func normalizeSQL(line string) string {
	return strings.TrimSpace(strings.TrimSuffix(line, ";"))
}
