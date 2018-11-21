package databases

type Benchmark struct {
	Name string
	Func func(int, int)
}
