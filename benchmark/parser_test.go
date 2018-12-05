package benchmark

import (
	"strings"
	"testing"
)

func TestParseScript(t *testing.T) {
	// arrange
	s := `
-- A sample script
-- {{.Iter}} and {{call .RandInt63}} will be replaced by the current iteration count and a random number.

\mode loop
BEGIN TRANSACTION;
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 
COMMIT;
/*
INSERT INTO dbbench_simple (id, balance) VALUES("IT'S A TRAP", 1);
*/

\mode loop
INSERT INTO dbbench_simple (id, balance) VALUES(1000, 1); -- inline comment
DELETE FROM dbbench_simple WHERE id = 1000; 

\mode once
INSERT INTO dbbench_simple (id, balance) VALUES(1000, 1);

DELETE FROM dbbench_simple WHERE id = 1000; 

\mode once
INSERT INTO dbbench_simple (id, balance) VALUES(1000, 1);

DELETE FROM dbbench_simple WHERE id = 1000; 

\mode loop
INSERT INTO dbbench_simple (id, balance) VALUES(1000, 1);
DELETE FROM dbbench_simple WHERE id = 1000;
`
	r := strings.NewReader(s)

	// act
	benchmarks := ParseScript(r)

	// assert
	got := len(benchmarks)
	want := 7
	if got != want {
		t.Errorf("got %v benchmarks, want %v", got, want)
	}

}
