package hysteria2

import (
	"os"
	"sync/atomic"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
)

var _ adapter.RuntimeUserManager = (*Inbound)(nil)

func (h *Inbound) ReplaceUsers(users []auth.User) error {
	if h.service == nil {
		return os.ErrClosed
	}
	if len(users) == 0 {
		return E.New("missing users")
	}
	seen := make(map[string]struct{}, len(users))
	names := make([]string, len(users))
	passwords := make([]string, len(users))
	ids := make([]int, len(users))
	for i, user := range users {
		if user.Username == "" || user.Password == "" {
			return E.New("user name and password are required")
		}
		if _, exists := seen[user.Username]; exists {
			return E.New("duplicate user name")
		}
		seen[user.Username] = struct{}{}
		names[i], passwords[i], ids[i] = user.Username, user.Password, i
	}
	h.service.UpdateUsers(ids, passwords)
	h.userNameList.Store(&names)
	return nil
}

func (h *Inbound) RuntimeUsers() []auth.User {
	names := h.userNameList.Load()
	if names == nil {
		return nil
	}
	users := make([]auth.User, len(*names))
	for i, name := range *names {
		users[i].Username = name
	}
	return users
}

var _ atomic.Pointer[[]string]
