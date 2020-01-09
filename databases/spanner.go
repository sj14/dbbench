package databases

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/spanner"
	"github.com/sj14/dbbench/benchmark"
	"google.golang.org/api/option"
)

// Spanner implements the bencher interface.
type Spanner struct {
	client     *spanner.Client
	ctx        context.Context
	instanceID string
}

/*
NewSpanner returns a new Google Cloud Spanner bencher.

spannerDatabase - A valid database name has the form projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID.
gcpCredentialsFile - Optional, path to file with needed GCP credentials to access Spanner. If left blank,
the default behavior of gcp libraries will be used, by assuming GOOGLE_APPLICATION_CREDENTIALS is set to the correct path
*/
func NewSpanner(projectID, instanceID, databaseID, gcpCredentialsFile string) *Spanner {
	ctx := context.Background()
	if projectID == "" {
		log.Fatalln("no projectID supplied to Spanner bencher")
	}
	if instanceID == "" {
		log.Fatalln("no instanceID supplied to Spanner bencher")
	}
	if databaseID == "" {
		log.Fatalln("no databaseID supplied to Spanner bencher")
	}

	gcpOpts := []option.ClientOption{}

	if gcpCredentialsFile != "" {
		gcpOpts = append(gcpOpts, option.WithCredentialsFile(gcpCredentialsFile))
	}

	database := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	client, err := spanner.NewClient(ctx, database, gcpOpts...)
	if err != nil {
		log.Fatalf("failed to open connection to spanner: %v", err)
	}

	return &Spanner{client, ctx, instanceID}
}

// Setup initializes the database for the benchmark.
func (s *Spanner) Setup() {}

// Cleanup removes all remaining benchmarking data.
func (s *Spanner) Cleanup() {}

// Benchmarks returns the individual benchmark functions for tspanner (not implemented).
func (s *Spanner) Benchmarks() (bb []benchmark.Benchmark) {
	log.Fatal("no built-in benchmarks for Spanner available yet, use your own script")
	return
}

// Exec executes the given statement on the database.
func (s *Spanner) Exec(stmt string) {
	_, err := s.client.ReadWriteTransaction(s.ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		txn.Query(ctx, spanner.NewStatement(stmt))
		return nil
	})
	if err != nil {
		log.Fatalf("%v: failed: %v\n", stmt, err)
	}
}
