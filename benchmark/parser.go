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
		names      []string        // queue of names, allow to set name before \mode, which might flush loop statement
		loopStmt   = ""            // loop stmt which might grow while parsing
		loopStart  = 1             // line the current loop mode started
		lineN      = 1             // current line number
	)

	// Helper function to determine the benchmark name.
	getName := func() string {
		if len(names) > 0 {
			if mode == TypeLoop {
				name := "(loop) " + names[0]
				return name
			}
			name := "(once) " + names[0]
			return name
		}
		switch mode {
		case TypeLoop:
			return fmt.Sprintf("(loop) line %v-%v", loopStart, lineN-1)
		case TypeOnce:
			return fmt.Sprintf("(once) line %v", lineN)
		}
		return "" // shouldn't happen
	}

	// Helper function to append a new loop benchmark
	flushLoop := func() {
		if loopStmt != "" {
			// was loop before, flush remaining loop statements
			loopStmt = strings.TrimSuffix(loopStmt, "\n")
			benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeLoop, Stmt: loopStmt})
			loopStmt = ""
			if len(names) > 0 {
				names = names[1:]
			}
		}
	}

	// Parse each line of the script file
	for ; scanner.Scan(); lineN++ {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		// Parse '\name' command. Set name and continue with next line.
		if strings.HasPrefix(line, "\\name ") {
			names = append(names, strings.TrimPrefix(line, "\\name "))
			continue
		}

		// Parse '\mode' command.
		if strings.HasPrefix(line, "\\mode ") {
			if strings.Contains(line, "once") {
				// once
				if mode == TypeLoop {
					flushLoop()
				}
				mode = TypeOnce
			} else if strings.Contains(line, "loop") {
				// loop
				flushLoop()
				mode = TypeLoop
				loopStart = lineN + 1
			} else {
				log.Fatalf("failed to parse mode, neither 'once' nor 'loop': %v", line)
			}

			// don't append '\mode' line
			continue
		}

		// Neither a '\mode' nor '\name' command line.
		// Append the line either as benchmark type once
		// or append line for loop benchmark.
		switch mode {
		case TypeOnce:
			// Once, append benchmark immediately.
			benchmarks = append(benchmarks, Benchmark{Name: getName(), Type: TypeOnce, Stmt: line})
			if len(names) > 0 {
				names = names[1:]
			}
		case TypeLoop:
			// Loop, but not finished yet, only append the line to the statement.
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
