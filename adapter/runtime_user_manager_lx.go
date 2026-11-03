package adapter

import "github.com/sagernet/sing/common/auth"

// RuntimeUserManager replaces credentials on a running server inbound.
type RuntimeUserManager interface {
	ReplaceUsers([]auth.User) error
	RuntimeUsers() []auth.User
}
