package db

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	log "github.com/sirupsen/logrus"
	"time"
)

var DB *badger.DB

func Load(cfg *config.Config) error {
	// It will be created if it doesn't exist.
	options := badger.DefaultOptions(cfg.DatabasePath)
	options.Logger = log.StandardLogger()

	db, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	return err
}

// StartValueLogGC starts a background goroutine that runs value log garbage collection
// periodically. It returns a function that can be called to stop the GC process.
func StartValueLogGC(ctx context.Context) func() {
	ticker := time.NewTicker(5 * time.Minute)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()

		// Run GC immediately on start
		runGC(DB)

		for {
			select {
			case <-ticker.C:
				runGC(DB)
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	// Return cleanup function
	return func() {
		ticker.Stop()
		<-done // Wait for goroutine to finish
	}
}

// runGC executes the value log garbage collection
func runGC(db *badger.DB) {
	// RunValueLogGC returns error when there's nothing to clean
	_ = db.RunValueLogGC(0.5)
	log.Debug("Ran Datbase GC")
}
