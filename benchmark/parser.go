package benchmark

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

// ParseScript parses a benchmark script and returns the benchmarks.
func ParseScript(r io.Reader) []Benchmark {
	s := bufio.NewScanner(r)
	benchmarks := []Benchmark{}

	mode := TypeLoop
	loopStmt := ""
	loopStart := 1
	lineN := 1
	for ; s.Scan(); lineN++ {
		line := strings.TrimSpace(s.Text())

		// skip comments and empty lines
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "\\mode") {
			if strings.Contains(line, "once") {
				// once
				if mode == TypeLoop {
					if loopStmt != "" {
						// was loop before, flush remaining loop statements
						benchmarks = append(benchmarks, Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: TypeLoop, Stmt: loopStmt})
						loopStmt = ""
					}
				}
				mode = TypeOnce
			} else if strings.Contains(line, "loop") {
				// loop
				if loopStmt != "" {
					// also was loop before, flush loop statements and start a new loop statement
					benchmarks = append(benchmarks, Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: TypeLoop, Stmt: loopStmt})
					loopStmt = ""
				}
				mode = TypeLoop
				loopStart = lineN + 1
			} else {
				log.Fatalf("failed to parse mode, neither 'once' nor 'loop': %v", line)
			}
			// don't append \mode line
			continue
		}

		switch mode {
		case TypeOnce:
			// Once, append benchmark immediately.
			benchmarks = append(benchmarks, Benchmark{Name: fmt.Sprintf("once: line %v", lineN), Type: TypeOnce, Stmt: line})
		case TypeLoop:
			// Loop, but not finished yet, append only line.
			loopStmt += line + "\n"
		}
	}

	// reached the end of the file, append remaining loop statements to benchmark
	if loopStmt != "" {
		loopStmt = strings.TrimSuffix(loopStmt, "\n")
		benchmarks = append(benchmarks, Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: TypeLoop, Stmt: loopStmt})
	}

	return benchmarks
}
