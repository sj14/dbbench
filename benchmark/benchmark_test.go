package benchmark

import (
	"testing"
	"text/template"
)

func TestBuildStmt(t *testing.T) {
	// arrange
	tmpl := template.New("test")
	tmpl.Parse("{{.Iter}} {{call .RandInt63}}")

	// act
	stmt := buildStmt(tmpl, 1337)

	// assert
	want := "1337 5577006791947779410"
	if stmt != want {
		t.Errorf("got statement %v, want %v", stmt, want)
	}
}
