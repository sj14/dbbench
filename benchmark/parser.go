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
	var (
		scanner    = bufio.NewScanner(r)
		benchmarks = []Benchmark{} // the result
		mode       = TypeLoop      // default mode is loop
		name       = ""            // name of the current benchmark, if empty use line numbers
		loopStmt   = ""            // loop stmt which might grow while parsing
		loopStart  = 1             // line the current loop mode started
		lineN      = 1             // current line number
	)

	// Helper function to determine the benchmark name.
	getName := func() string {
		if name != "" {
			return name
		}
		switch mode {
		case TypeLoop:
			return fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1)
		case TypeOnce:
			return fmt.Sprintf("once: line %v", lineN)
		}
		return "" // shouldn't happen
	}

	// Scan each line of the file
	for ; scanner.Scan(); lineN++ {
		line := strings.TrimSpace(scanner.Text())

		// skip comments and empty lines
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		// Found name command. Set name and continue with next line.
		if strings.HasPrefix(line, "\\name ") {
			name = strings.TrimPrefix(line, "\\name ")
			continue
		}

		if strings.HasPrefix(line, "\\mode ") {
			if strings.Contains(line, "once") {
				// once
				if mode == TypeLoop {
					if loopStmt != "" {
						// was loop before, flush remaining loop statements
						benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeLoop, Stmt: loopStmt})
						loopStmt = ""
					}
				}
				mode = TypeOnce
			} else if strings.Contains(line, "loop") {
				// loop
				if loopStmt != "" {
					// also was loop before, flush loop statements and start a new loop statement
					benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeLoop, Stmt: loopStmt})
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
			benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeOnce, Stmt: line})
		case TypeLoop:
			// Loop, but not finished yet, append only line.
			loopStmt += line + "\n"
		}
	}

	// reached the end of the file, append remaining loop statements to benchmark
	if loopStmt != "" {
		loopStmt = strings.TrimSuffix(loopStmt, "\n")
		benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeLoop, Stmt: loopStmt})
	}

	return benchmarks
}
