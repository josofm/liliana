package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	entity "github.com/josofm/liliana/internal/entity/group"
	user "github.com/josofm/liliana/internal/entity/user"
	repository "github.com/josofm/liliana/internal/repository/group"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	service "github.com/josofm/liliana/internal/service/group"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroupHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := userRepo.NewInMemoryRepo()
	require.NoError(t, users.Create(&user.User{Name: "Owner"}))
	require.NoError(t, users.Create(&user.User{Name: "Member"}))
	repo := repository.NewInMemoryRepo()
	s := service.NewService(repo, users)
	request := func(actor int64, method, path, body string) *httptest.ResponseRecorder {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			if actor > 0 {
				c.Set("user_id", actor)
			}
		})
		setupGroupRoutes(r, s)
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, request(0, "GET", "/groups/", "").Code)
	w := request(1, "POST", "/groups/", `{"name":"Group","owner_id":2,"role":"member"}`)
	require.Equal(t, 201, w.Code)
	var g entity.Group
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))
	require.Equal(t, int64(1), g.OwnerID)
	require.Equal(t, 403, request(2, "GET", "/groups/1", "").Code)
	require.Equal(t, 200, request(1, "GET", "/groups/1/members", "").Code)
	require.NoError(t, repo.AddMember(1, 2))
	for _, tt := range []struct {
		actor              int64
		method, path, body string
		status             int
	}{
		{1, "POST", "/groups/", `{"name":"  "}`, 400},
		{1, "POST", "/groups/", `{`, 400},
		{1, "GET", "/groups/bad", "", 400},
		{1, "GET", "/groups/999", "", 404},
		{2, "GET", "/groups/1", "", 200},
		{2, "PATCH", "/groups/1", `{"name":"Changed"}`, 403},
		{1, "PATCH", "/groups/1", `{"description":"Updated","owner_id":2}`, 200},
		{1, "PATCH", "/groups/1", `{}`, 400},
		{1, "DELETE", "/groups/1/members/1", "", 409},
		{2, "DELETE", "/groups/1/members/1", "", 403},
		{2, "DELETE", "/groups/1/members/2", "", 204},
		{2, "GET", "/groups/1", "", 403},
	} {
		t.Run(tt.method+tt.path+tt.body, func(t *testing.T) {
			w := request(tt.actor, tt.method, tt.path, tt.body)
			require.Equal(t, tt.status, w.Code, w.Body.String())
		})
	}
	w = request(2, http.MethodGet, "/groups/", "")
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `[]`, w.Body.String())
}
