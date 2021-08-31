package indexer_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/indexer"
)

func Test_Indexer_001(t *testing.T) {
	tmppath, err := os.MkdirTemp("", "indexer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmppath)
	if indexer, err := indexer.New("test", tmppath, make(chan<- server.Event)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(indexer)
	}

}

func Test_Indexer_002(t *testing.T) {
	var wg sync.WaitGroup
	tmppath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	indexer, err := indexer.New("test", tmppath, make(chan<- server.Event))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		go indexer.Walk(ctx)
		if err := indexer.Run(ctx); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}
