package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/config"
	entity "github.com/josofm/liliana/internal/entity/group"
	user "github.com/josofm/liliana/internal/entity/user"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	groupRepo "github.com/josofm/liliana/internal/repository/group"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/internal/service/auth"
	"github.com/josofm/liliana/pkg/logger"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInvitationHTTPFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := userRepo.NewInMemoryRepo()
	groups := groupRepo.NewInMemoryRepo()
	jwt := auth.NewJWTService(auth.JWTConfig{SecretKey: "invite-test", AccessExpiry: time.Hour, RefreshExpiry: time.Hour})
	tokens := map[int64]string{}
	for _, u := range []*user.User{{Name: "Owner", Email: "owner@example.com"}, {Name: "Marina", Email: "sea@example.com"}, {Name: "Other", Email: "other@example.com"}} {
		require.NoError(t, users.Create(u))
		pair, err := jwt.GenerateTokenPair(u.ID, u.Email)
		require.NoError(t, err)
		tokens[u.ID] = pair.AccessToken
	}
	router := gin.New()
	NewRouter(router, logger.New("error"), users, deckRepo.NewInMemoryRepo(), groups, &config.Config{JWT: config.JWTConfig{SecretKey: "invite-test", AccessExpiry: time.Hour, RefreshExpiry: time.Hour}})
	request := func(actor int64, method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token := tokens[actor]; token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, status, w.Code, w.Body.String())
		return w
	}
	request(0, "GET", "/users/search?q=ma", "", 401)
	for _, q := range []string{"", "a", "%20%20", strings.Repeat("a", 101)} {
		request(1, "GET", "/users/search?q="+q, "", 400)
	}
	nameOnly := request(1, "GET", "/users/search?q=Marina", "", 200)
	require.JSONEq(t, `[]`, nameOnly.Body.String())
	for _, q := range []string{"sEa", "SEA%40example.com"} {
		w := request(1, "GET", "/users/search?q="+q, "", 200)
		require.JSONEq(t, `[{"id":2,"name":"Marina","email":"sea@example.com"}]`, w.Body.String())
	}
	request(1, "POST", "/groups/", `{"name":"Commander Night"}`, 201)
	request(0, "GET", "/group-invitations/", "", 401)
	request(0, "POST", "/group-invitations/1/accept", "", 401)
	request(1, "POST", "/groups/1/invitations", `{`, 400)
	request(1, "POST", "/groups/1/invitations", `{}`, 400)
	request(1, "POST", "/groups/1/invitations", `{"invited_player_id":999}`, 404)
	request(1, "POST", "/groups/1/invitations", `{"invited_player_id":1}`, 409)
	request(3, "POST", "/groups/1/invitations", `{"invited_player_id":2}`, 403)
	w := request(1, "POST", "/groups/1/invitations", `{"invited_player_id":2,"invited_by_player_id":3,"status":"accepted"}`, 201)
	var i entity.Invitation
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &i))
	require.Equal(t, int64(1), i.InvitedByPlayerID)
	require.Equal(t, entity.InvitationPending, i.Status)
	request(1, "POST", "/groups/1/invitations", `{"invited_player_id":2}`, 409)
	request(2, "GET", "/groups/1", "", 403)
	request(2, "GET", "/groups/1/invitations", "", 403)
	w = request(3, "GET", "/group-invitations/", "", 200)
	require.JSONEq(t, `[]`, w.Body.String())
	w = request(2, "GET", "/group-invitations/", "", 200)
	require.Contains(t, w.Body.String(), "Commander Night")
	w = request(1, "GET", "/groups/1/invitations", "", 200)
	require.Contains(t, w.Body.String(), `"invited_player_id":2`)
	request(1, "POST", "/group-invitations/1/accept", "", 403)
	request(3, "POST", "/group-invitations/1/decline", "", 403)
	request(2, "DELETE", "/group-invitations/1", "", 403)
	request(2, "POST", "/group-invitations/invalid/accept", "", 400)
	request(2, "POST", "/group-invitations/999/accept", "", 404)
	request(2, "POST", "/group-invitations/1/accept", "", 200)
	request(2, "GET", "/groups/1", "", 200)
	request(2, "POST", "/group-invitations/1/accept", "", 409)
	request(1, "DELETE", "/group-invitations/1", "", 409)
	request(2, "POST", "/groups/1/invitations", `{"invited_player_id":3}`, 403)
	w = request(2, "GET", "/group-invitations/", "", 200)
	require.JSONEq(t, `[]`, w.Body.String())
	request(1, "POST", "/groups/1/invitations", `{"invited_player_id":3}`, 201)
	request(3, "POST", "/group-invitations/2/decline", "", 200)
	request(3, "GET", "/groups/1", "", 403)
	request(1, "POST", "/groups/1/invitations", `{"invited_player_id":3}`, 201)
	request(1, "DELETE", "/group-invitations/3", "", 204)
	request(3, "POST", "/group-invitations/3/accept", "", 409)
	w = request(1, "GET", "/groups/1/invitations", "", 200)
	require.JSONEq(t, `[]`, w.Body.String())
}
