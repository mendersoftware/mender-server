package mongo

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestLock(t *testing.T) {
	mongoURL := os.Getenv("TEST_MONGO_URL")
	if testing.Short() {
		t.Skip()
	} else if mongoURL == "" {
		t.Skip("env: TEST_MONGO_URL not set")
	}
	t.Parallel()
	ctx := context.Background()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Errorf("failed to setup mongodb client: %s", err.Error())
		t.FailNow()
	}
	dbRepl := strings.NewReplacer("/", "").Replace

	t.Run("basic functionality", func(t *testing.T) {
		db := c.Database(dbRepl(t.Name()))
		defer db.Drop(ctx)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
		defer cancel()

		lock := NewLock(db, "locks", "l1")
		ok, err := lock.TryLock(ctx)
		if err != nil {
			t.Errorf("unexpected error aquiring lock: %s", err.Error())
			t.FailNow()
		}
		if !ok {
			t.Errorf("failed to aquire lock on empty database")
			t.FailNow()
		}
		anotherLock := NewLock(db, "locks", "l1")
		ok, err = anotherLock.TryLock(ctx)
		if err != nil {
			t.Errorf("unexpected error aquiring lock again")
			t.FailNow()
		} else if ok {
			t.Error("lock should already be aquired")
		}
		err = lock.Unlock(ctx)
		if err != nil {
			t.Errorf("unexpected error releasing lock: %s", err.Error())
		}
	})
	t.Run("concurrency stress test", func(t *testing.T) {
		db := c.Database(dbRepl(t.Name()))
		defer db.Drop(ctx)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer cancel()
		lock := NewLock(db, "locks", "l1")
		wgUp := new(sync.WaitGroup)
		wgUp.Add(100)
		wgDown := new(sync.WaitGroup)
		wgDown.Add(100)
		barrier := make(chan struct{})
		var c uint64
		for range 100 {
			go func() {
				wgUp.Done()
				<-barrier
				ok, err := lock.TryLock(ctx)
				if err != nil {
					t.Errorf("failed to call TryLock: %s", err.Error())
				}
				if ok {
					atomic.AddUint64(&c, 1)
				}
				wgDown.Done()
			}()
		}
		wgUp.Wait()
		close(barrier)
		wgDown.Wait()
		if c == 0 {
			t.Error("no go routine aquired the lock")
		} else if c > 1 {
			t.Error("more than one go routine aquired the lock")
		}
	})
}
