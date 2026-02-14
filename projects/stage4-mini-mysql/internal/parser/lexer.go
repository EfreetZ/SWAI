package parser

import "strings"

// Tokenize 将 SQL 粗粒度分词。
func Tokenize(sql string) []string {
	sql = strings.TrimSpace(sql)
	sql = strings.ReplaceAll(sql, ",", " ")
	sql = strings.ReplaceAll(sql, "(", " ")
	sql = strings.ReplaceAll(sql, ")", " ")
	sql = strings.ReplaceAll(sql, "=", " = ")
	sql = strings.ReplaceAll(sql, ";", "")
	parts := strings.Fields(sql)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
