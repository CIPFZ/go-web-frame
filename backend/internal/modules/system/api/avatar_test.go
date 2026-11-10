package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
)

type avatarUsers struct {
	service.IUserService
	saved string
}

func (u *avatarUsers) UpdateAvatar(_ context.Context, _ uuid.UUID, avatar string) error {
	u.saved = avatar
	return nil
}

func TestSetPresetAvatarAcceptsOnlyBuiltInAvatars(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := &avatarUsers{}
	api := &UserApi{userService: users}
	call := func(body string) map[string]any {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/sys/user/avatar/preset", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.SetPresetAvatar(c)
		var out map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
		return out
	}
	for _, bad := range []string{`{"avatar":"https://example.com/x.png"}`, `{"avatar":"/avatars/../etc/passwd"}`, `{}`} {
		require.NotEqual(t, float64(0), call(bad)["code"], bad)
	}
	require.Empty(t, users.saved)
	require.Equal(t, float64(0), call(`{"avatar":"/avatars/animal-fox.svg"}`)["code"])
	require.Equal(t, "/avatars/animal-fox.svg", users.saved)
}
