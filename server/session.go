package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soooooooon/gin-contrib/gin_helper"
	"github.com/soooooooon/rock/models"
)

const (
	sessionContext = "user_session_id_key"
)

func LoginRequired(c *gin.Context) {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer") {
		token := strings.TrimLeft(auth, "Bearer")
		token = strings.TrimSpace(token)

		s, err := models.SessionWithToken(c.Request.Context(), token)
		if err != nil {
			gin_helper.FailError(c, loginRequired, err.Error())
			return
		}

		c.Set(sessionContext, s)
		return
	}

	gin_helper.FailError(c, loginRequired, "Bearer token not found in header")
}

func ExtractSession(c *gin.Context) *models.Session {
	return c.MustGet(sessionContext).(*models.Session)
}

func ExtractMixinUserId(c *gin.Context) string {
	return ExtractSession(c).UserId
}

func queryMyProfile(c *gin.Context) {
	id := ExtractMixinUserId(c)
	ctx := c.Request.Context()
	user, err := models.GetUserWithMixinID(ctx, id)
	if err != nil {
		gin_helper.FailError(c, err)
		return
	}

	gin_helper.OK(c, user)
}
