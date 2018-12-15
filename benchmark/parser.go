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
		curBench   = Benchmark{Type: TypeLoop, Parallel: false}
		loopStart  = 1 // line the current loop mode started
		lineN      = 1 // current line number
	)

	// Helper function to determine the benchmark name.
	getName := func() string {
		switch curBench.Type {
		case TypeLoop:
			if curBench.Name != "" {
				return "(loop) " + curBench.Name
			}
			return fmt.Sprintf("(loop) line %v-%v", loopStart, lineN-1)
		case TypeOnce:
			if curBench.Name != "" {
				return "(once) " + curBench.Name
			}
			return fmt.Sprintf("(once) line %v", lineN)
		}
		return "" // shouldn't happen
	}

	// Helper function to append a new loop benchmark
	flushLoop := func() {
		if curBench.Stmt != "" {
			curBench.Stmt = strings.TrimSuffix(curBench.Stmt, "\n")
			curBench.Name = getName()
			benchmarks = append(benchmarks, curBench)

			// Start new empty benchmark
			curBench = Benchmark{}
		}
	}

	// Parse each line of the script file
	for ; scanner.Scan(); lineN++ {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		// Parse '\benchmark' command.
		if strings.HasPrefix(line, "\\benchmark ") {
			tokens := strings.Split(line, " ")

			// remove '\benchmark' entry from tokens
			tokens = tokens[1:]

			if len(tokens) <= 0 {
				// Line does only consist of the token '\benchmark', we need more info.
				log.Fatalf("failed to parse \\benchmark line, too few arguments")
			}

			switch tokens[0] {
			case "once":
				if curBench.Type == TypeLoop {
					flushLoop()
				}
				curBench.Type = TypeOnce
			case "loop":
				flushLoop()
				curBench.Type = TypeLoop
				loopStart = lineN + 1
			default:
				log.Fatalf("failed to parse mode, neither 'once' nor 'loop': %v", line)
			}
			// remove the mode token from the tokens
			tokens = tokens[1:]

			// Parse remaining tokens
			for _, t := range tokens {
				// reminder: can't change 'tokens' inside the range, e.g. 'cutting' with tokens[2:]
				// so we have to iterate even the token after \name, which could be skipped otherwise.
				switch t {
				case "\\parallel":
					curBench.Parallel = true
				case "\\name":
					if len(tokens) < 2 {
						log.Fatalf("missing name after \\name token")
					}
					curBench.Name = tokens[1]
				}
			}

			// don't append '\benchmark' line
			continue
		}

		// Neither a '\benchmark' nor '\name' command line.
		// Append the line either as benchmark type once
		// or append line for loop benchmark.
		switch curBench.Type {
		case TypeOnce:
			// Once, append benchmark immediately.
			curBench.Type = TypeOnce
			curBench.Name = getName()
			curBench.Stmt = line
			benchmarks = append(benchmarks, curBench)
			// As long as there is no mode change, keep it TypeOnce, which is the non-default mode.
			curBench = Benchmark{Type: TypeOnce}
		case TypeLoop:
			// Loop, but not finished yet, only append the line to the statement.
			curBench.Stmt += line + "\n"
		}
	}

	// reached the end of the file, append remaining loop statements to benchmark
	if curBench.Stmt != "" {
		curBench.Stmt = strings.TrimSuffix(curBench.Stmt, "\n")
		curBench.Name = getName()
		benchmarks = append(benchmarks, curBench)
	}

	return benchmarks
}
