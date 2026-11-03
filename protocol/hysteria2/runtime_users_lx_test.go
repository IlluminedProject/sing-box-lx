package hysteria2

import (
	"testing"

	"github.com/sagernet/sing/common/auth"
)

// lx: runtime-management — snapshot semantics without a QUIC listener. The
// service handle stays nil, so only the failure path and the stored snapshot
// are observable here; full-stack coverage lives in the integration tests.
func TestRuntimeUserSnapshotSemantics(t *testing.T) {
	inbound := &Inbound{}
	names := []string{"alice"}
	inbound.userNameList.Store(&names)
	users := inbound.RuntimeUsers()
	if len(users) != 1 || users[0].Username != "alice" || users[0].Password != "" {
		t.Fatalf("unexpected public users: %+v", users)
	}
	if err := inbound.ReplaceUsers(nil); err == nil {
		t.Fatal("empty replacement accepted")
	}
	if err := inbound.ReplaceUsers([]auth.User{{Username: "alice", Password: "x"}}); err == nil {
		t.Fatal("nil service replacement accepted")
	}
	if got := inbound.RuntimeUsers(); len(got) != 1 || got[0].Username != "alice" {
		t.Fatalf("failed replacement changed snapshot: %+v", got)
	}
}
