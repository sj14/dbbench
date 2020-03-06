package benchmark

import (
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"

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
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt63}}"))

	// act
	stmt := buildStmt(tmpl, 1337)

	// assert
	want := "1337 5577006791947779410"
	if stmt != want {
		t.Errorf("got statement %v, want %v", stmt, want)
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		description string
		givenType   BenchType
	}{
		{
			description: "loop",
			givenType:   TypeLoop,
		},
		{
			description: "once",
			givenType:   TypeOnce,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			// arrange
			bencher := &mockedBencher{}
			bencher.On("Exec", mock.Anything)

			iter := 13
			threads := 5
			bLoop := Benchmark{Name: "test", Type: tt.givenType, Stmt: "NONE"}

			// act
			Run(bencher, bLoop, iter, threads)

			// assert
			switch tt.givenType {
			case TypeLoop:
				bencher.AssertNumberOfCalls(t, "Exec", iter)
			case TypeOnce:
				bencher.AssertNumberOfCalls(t, "Exec", 1)
			}
		})
	}
}
func TestLoop(t *testing.T) {
	// arrange
	bencher := &mockedBencher{}
	bencher.On("Exec", mock.Anything)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt63}}"))

	executor := bencherExecutor{
		result: Result{
			Start: time.Now(),
		},
	}

	// act
	executor.loop(bencher, tmpl, 17, 5)

	// assert
	bencher.AssertNumberOfCalls(t, "Exec", 17)
}

func TestOnce(t *testing.T) {
	// arrange
	bencher := &mockedBencher{}
	bencher.On("Exec", mock.Anything)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt63}}"))

	executor := bencherExecutor{
		result: Result{
			Start: time.Now(),
		},
	}

	// act
	executor.once(bencher, tmpl)

	// assert
	bencher.AssertNumberOfCalls(t, "Exec", 1)
}

func TestResults(t *testing.T) {
	// arrange
	bencher := &mockedBencher{}
	bencher.On("Exec", mock.Anything)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt63}}"))

	executor := bencherExecutor{
		result: Result{
			Start: time.Now(),
		},
	}

	// act
	executor.once(bencher, tmpl)

	assert.Equal(t, uint64(1), executor.result.TotalExecutionCount)

	assert.Equal(t, executor.result.TotalExecutionTime, executor.result.Avg())
}
