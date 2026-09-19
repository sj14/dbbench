package benchmark

import (
	"errors"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedBencher struct {
	mock.Mock
}

func TestRunReturnsExecutionError(t *testing.T) {
	bencher := &mockedBencher{}
	want := errors.New("execution failed")
	bencher.On("Exec", mock.Anything).Return(want)

	result, err := Run(bencher, Benchmark{Name: "failure", Type: TypeLoop, Stmt: "NONE"}, 10, 2)

	assert.ErrorIs(t, err, want)
	assert.Zero(t, result.TotalExecutionCount)
}

func (b *mockedBencher) Benchmarks() []Benchmark { return []Benchmark{} }
func (b *mockedBencher) Setup()                  {}
func (b *mockedBencher) Cleanup()                {}
func (b *mockedBencher) Exec(s string) error     { return b.Called(s).Error(0) }

func TestBuildStmt(t *testing.T) {
	// arrange
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} test"))

	// act
	stmt := buildStmt(tmpl, 1337)

	// assert
	want := "1337 test"
	if stmt != want {
		t.Errorf("got statement %v, want %v", stmt, want)
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		description string
		givenType   BenchType
		parallel    bool
	}{
		{
			description: "loop",
			givenType:   TypeLoop,
		},
		{
			description: "once",
			givenType:   TypeOnce,
		},
		{
			description: "parallel loop waits for completion",
			givenType:   TypeLoop,
			parallel:    true,
		},
		{
			description: "parallel once waits for completion",
			givenType:   TypeOnce,
			parallel:    true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			// arrange
			bencher := &mockedBencher{}
			bencher.On("Exec", mock.Anything).Return(nil)

			iter := 13
			threads := 5
			bLoop := Benchmark{Name: "test", Type: tt.givenType, Parallel: tt.parallel, Stmt: "NONE"}

			// act
			_, err := Run(bencher, bLoop, iter, threads)

			// assert
			assert.NoError(t, err)
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
	bencher.On("Exec", mock.Anything).Return(nil)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt64}}"))

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
	bencher.On("Exec", mock.Anything).Return(nil)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt64}}"))

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
	bencher.On("Exec", mock.Anything).Return(nil)
	tmpl := template.Must(template.New("test").Parse("{{.Iter}} {{call .RandInt64}}"))

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
