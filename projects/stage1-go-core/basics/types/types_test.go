package types

import (
	"reflect"
	"testing"
)

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Env
		wantErr bool
	}{
		{name: "development", input: "dev", want: EnvDevelopment},
		{name: "testing", input: " testing ", want: EnvTesting},
		{name: "production", input: "PROD", want: EnvProduction},
		{name: "invalid", input: "uat", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnv(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseEnv() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEnv() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSwapPair(t *testing.T) {
	pair := Pair[string, int]{Left: "left", Right: 42}
	swapped := SwapPair(pair)
	if swapped.Left != 42 || swapped.Right != "left" {
		t.Fatalf("SwapPair() = %+v, want Left=42 Right=left", swapped)
	}
}

func TestParseIntDefault(t *testing.T) {
	if got, err := ParseIntDefault("", 10); err != nil || got != 10 {
		t.Fatalf("ParseIntDefault(empty) = (%d, %v), want (10, nil)", got, err)
	}

	if got, err := ParseIntDefault("21", 10); err != nil || got != 21 {
		t.Fatalf("ParseIntDefault(21) = (%d, %v), want (21, nil)", got, err)
	}

	if got, err := ParseIntDefault("abc", 10); err == nil || got != 10 {
		t.Fatalf("ParseIntDefault(abc) = (%d, %v), want (10, error)", got, err)
	}
}

func TestFilter(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := Filter(items, func(v int) bool {
		return v%2 == 0
	})
	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Filter() = %v, want %v", got, want)
	}
}

func BenchmarkFilter(b *testing.B) {
	items := make([]int, 0, 1000)
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Filter(items, func(v int) bool {
			return v%3 == 0
		})
	}
}
