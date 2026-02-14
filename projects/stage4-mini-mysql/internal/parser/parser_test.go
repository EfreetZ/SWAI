package parser

import "testing"

func TestParseStatements(t *testing.T) {
	cases := []string{
		"CREATE TABLE kv;",
		"INSERT INTO kv k1 v1;",
		"SELECT * FROM kv WHERE key = k1;",
		"SELECT RANGE FROM kv a z;",
		"BEGIN;",
		"COMMIT;",
		"ROLLBACK;",
	}
	for _, sql := range cases {
		if _, err := Parse(sql); err != nil {
			t.Fatalf("Parse(%q) error = %v", sql, err)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("DROP TABLE kv;"); err == nil {
		t.Fatal("Parse(invalid) should fail")
	}
}
