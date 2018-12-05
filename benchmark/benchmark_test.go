package benchmark

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/mock"
)

type mockedBencher struct {
	mock.Mock
}

func (b *mockedBencher) Benchmarks() []Benchmark { return []Benchmark{} }
func (b *mockedBencher) Setup()                  {}
func (b *mockedBencher) Cleanup()                {}
func (b *mockedBencher) Exec(s string)           { _ = b.Called(s) }

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

func TestLoop(t *testing.T) {
	// arrange
	bencher := &mockedBencher{}
	bencher.On("Exec", mock.Anything)

	tmpl := template.New("test")
	tmpl.Parse("{{.Iter}} {{call .RandInt63}}")

	// act
	loop(bencher, tmpl, 10, 5)

	// assert
	bencher.AssertNumberOfCalls(t, "Exec", 10)
}

func TestOnce(t *testing.T) {
	// arrange
	bencher := &mockedBencher{}
	bencher.On("Exec", mock.Anything)

	tmpl := template.New("test")
	tmpl.Parse("{{.Iter}} {{call .RandInt63}}")

	// act
	once(bencher, tmpl)

	// assert
	bencher.AssertNumberOfCalls(t, "Exec", 1)
}
