package clashapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common/auth"
	"github.com/go-chi/render"
)

// lx: runtime-management — REST seam coverage without a running box.

type stubUserInbound struct {
	adapter.Inbound
	users []auth.User
}

func (s *stubUserInbound) ReplaceUsers(users []auth.User) error {
	s.users = users
	return nil
}

func (s *stubUserInbound) RuntimeUsers() []auth.User {
	return s.users
}

func renderJSON(writer http.ResponseWriter, payload render.M) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(payload)
}

func newRuntimeTestRouter(userManager adapter.RuntimeUserManager) http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /inbounds/{tag}/users", func(writer http.ResponseWriter, request *http.Request) {
		renderJSON(writer, render.M{"users": publicRuntimeUsers(userManager.RuntimeUsers())})
	})
	handler.HandleFunc("PUT /inbounds/{tag}/users", func(writer http.ResponseWriter, request *http.Request) {
		var payload runtimeUserRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		users := make([]auth.User, len(payload.Users))
		for i, user := range payload.Users {
			users[i] = auth.User{Username: user.Username, Password: user.Password}
		}
		if err := userManager.ReplaceUsers(users); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		renderJSON(writer, render.M{"users": publicRuntimeUsers(userManager.RuntimeUsers())})
	})
	return handler
}

func TestRuntimeUsersHappyPath(t *testing.T) {
	inbound := &stubUserInbound{}
	router := newRuntimeTestRouter(inbound)
	payload := `{"users":[{"username":"alice","password":"secret"}]}`
	request := httptest.NewRequest(http.MethodPut, "/inbounds/hy2/users", strings.NewReader(payload))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(inbound.users) != 1 || inbound.users[0].Username != "alice" || inbound.users[0].Password != "secret" {
		t.Fatalf("replacement not applied: %+v", inbound.users)
	}
	if strings.Contains(recorder.Body.String(), "secret") {
		t.Fatal("password leaked in response")
	}
}

func TestPublicRuntimeUsersRedactsPassword(t *testing.T) {
	users := publicRuntimeUsers([]auth.User{{Username: "alice", Password: "secret"}})
	if len(users) != 1 {
		t.Fatalf("unexpected result: %+v", users)
	}
	if _, hasPassword := users[0]["password"]; hasPassword {
		t.Fatal("password must not be exposed")
	}
}
