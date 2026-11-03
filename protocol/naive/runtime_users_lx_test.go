package naive

import (
	"sync"
	"testing"

	"github.com/sagernet/sing/common/auth"
)

func TestUserTableReplacement(t *testing.T) {
	table := newUserTable([]auth.User{{Username: "alice", Password: "old"}})
	if !table.verify("alice", "old") {
		t.Fatal("initial credentials refused")
	}
	table.replace([]auth.User{{Username: "alice", Password: "new"}})
	if table.verify("alice", "old") || !table.verify("alice", "new") {
		t.Fatal("replacement did not take effect atomically")
	}
}

func TestUserTableConcurrentReplacement(t *testing.T) {
	table := newUserTable([]auth.User{{Username: "alice", Password: "0"}})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 1000; j++ {
				table.verify("alice", "0")
				table.verify("alice", "1")
			}
		}()
	}
	for i := 0; i < 1000; i++ {
		password := "0"
		if i%2 != 0 {
			password = "1"
		}
		table.replace([]auth.User{{Username: "alice", Password: password}})
	}
	group.Wait()
}
