package clashapi

import (
	"net/http"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// lx:begin runtime-management
// Runtime management API (SPEC 019/152): live replacement of inbound users and
// WireGuard peers without restarting the instance. Inbounds/endpoints opt in by
// implementing adapter.RuntimeUserManager / adapter.RuntimePeerManager.

type runtimeUserRequest struct {
	Users []runtimeUser `json:"users"`
}

type runtimeUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type runtimePeersRequest struct {
	Peers []option.WireGuardPeer `json:"peers"`
}

func runtimeRouter(server *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/inbounds/{tag}/users", server.getRuntimeInboundUsers)
	r.Put("/inbounds/{tag}/users", server.replaceRuntimeInboundUsers)
	r.Get("/endpoints/{tag}/peers", server.getRuntimeEndpointPeers)
	r.Put("/endpoints/{tag}/peers", server.replaceRuntimeEndpointPeers)
	return r
}

func findRuntimeUserManager(server *Server, tag string) (adapter.RuntimeUserManager, error) {
	inboundManager := service.FromContext[adapter.InboundManager](server.ctx)
	if inboundManager == nil {
		return nil, E.New("inbound manager is unavailable")
	}
	inbound, found := inboundManager.Get(tag)
	if !found {
		return nil, E.New("inbound not found: ", tag)
	}
	manager, ok := inbound.(adapter.RuntimeUserManager)
	if !ok {
		return nil, E.New("inbound does not support runtime user management")
	}
	return manager, nil
}

func findRuntimePeerManager(server *Server, tag string) (adapter.RuntimePeerManager, error) {
	if server.endpoint == nil {
		return nil, E.New("endpoint manager is unavailable")
	}
	endpoint, found := server.endpoint.Get(tag)
	if !found {
		return nil, E.New("endpoint not found: ", tag)
	}
	manager, ok := endpoint.(adapter.RuntimePeerManager)
	if !ok {
		return nil, E.New("endpoint does not support runtime peer management")
	}
	return manager, nil
}

func (s *Server) getRuntimeInboundUsers(writer http.ResponseWriter, request *http.Request) {
	tag := chi.URLParam(request, "tag")
	manager, err := findRuntimeUserManager(s, tag)
	if err != nil {
		render.Status(request, http.StatusNotFound)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	render.JSON(writer, request, render.M{"users": publicRuntimeUsers(manager.RuntimeUsers())})
}

func (s *Server) replaceRuntimeInboundUsers(writer http.ResponseWriter, request *http.Request) {
	tag := chi.URLParam(request, "tag")
	manager, err := findRuntimeUserManager(s, tag)
	if err != nil {
		render.Status(request, http.StatusNotFound)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	var payload runtimeUserRequest
	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		render.Status(request, http.StatusBadRequest)
		render.JSON(writer, request, render.M{"error": "invalid payload: " + err.Error()})
		return
	}
	users := make([]auth.User, len(payload.Users))
	for i, user := range payload.Users {
		users[i] = auth.User{Username: user.Username, Password: user.Password}
	}
	if err = manager.ReplaceUsers(users); err != nil {
		render.Status(request, http.StatusBadRequest)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	render.JSON(writer, request, render.M{"users": publicRuntimeUsers(manager.RuntimeUsers())})
}

func (s *Server) getRuntimeEndpointPeers(writer http.ResponseWriter, request *http.Request) {
	tag := chi.URLParam(request, "tag")
	manager, err := findRuntimePeerManager(s, tag)
	if err != nil {
		render.Status(request, http.StatusNotFound)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	render.JSON(writer, request, render.M{"peers": manager.RuntimePeers()})
}

func (s *Server) replaceRuntimeEndpointPeers(writer http.ResponseWriter, request *http.Request) {
	tag := chi.URLParam(request, "tag")
	manager, err := findRuntimePeerManager(s, tag)
	if err != nil {
		render.Status(request, http.StatusNotFound)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	var payload runtimePeersRequest
	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		render.Status(request, http.StatusBadRequest)
		render.JSON(writer, request, render.M{"error": "invalid payload: " + err.Error()})
		return
	}
	if err = manager.ReplacePeers(payload.Peers); err != nil {
		render.Status(request, http.StatusBadRequest)
		render.JSON(writer, request, render.M{"error": err.Error()})
		return
	}
	render.JSON(writer, request, render.M{"peers": manager.RuntimePeers()})
}

func publicRuntimeUsers(users []auth.User) []render.M {
	result := make([]render.M, 0, len(users))
	for _, user := range users {
		result = append(result, render.M{"username": user.Username})
	}
	return result
}
// lx:end runtime-management
