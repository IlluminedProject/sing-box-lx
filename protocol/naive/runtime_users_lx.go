package naive

import (
	"os"
	"sync/atomic"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
)

var _ adapter.RuntimeUserManager = (*Inbound)(nil)

type userTable struct {
	authenticator atomic.Pointer[auth.Authenticator]
	users         atomic.Pointer[[]auth.User]
}

func newUserTable(users []auth.User) *userTable {
	table := new(userTable)
	table.replace(users)
	return table
}

func (t *userTable) replace(users []auth.User) {
	copyUsers := append([]auth.User(nil), users...)
	t.authenticator.Store(auth.NewAuthenticator(copyUsers))
	t.users.Store(&copyUsers)
}

func (t *userTable) verify(username, password string) bool {
	authenticator := t.authenticator.Load()
	return authenticator != nil && authenticator.Verify(username, password)
}

func (n *Inbound) ReplaceUsers(users []auth.User) error {
	if n.ctx.Err() != nil {
		return os.ErrClosed
	}
	if len(users) == 0 {
		return E.New("missing users")
	}
	seen := make(map[string]struct{}, len(users))
	for _, user := range users {
		if user.Username == "" || user.Password == "" {
			return E.New("user name and password are required")
		}
		if _, exists := seen[user.Username]; exists {
			return E.New("duplicate user name")
		}
		seen[user.Username] = struct{}{}
	}
	n.users.replace(users)
	return nil
}

func (n *Inbound) RuntimeUsers() []auth.User {
	users := n.users.users.Load()
	if users == nil {
		return nil
	}
	return append([]auth.User(nil), (*users)...)
}
