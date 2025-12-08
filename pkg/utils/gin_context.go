package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxKeyUserUUID    = "userUUID"
	ctxKeyUserID      = "userId"
	ctxKeyAuthorityID = "authorityId"
	ctxKeyClaims      = "claims"
)

// GetUserUUID 从 Context 中获取用户 UUID
func GetUserUUID(c *gin.Context) uuid.UUID {
	if v, exists := c.Get(ctxKeyUserUUID); exists {
		if u, ok := v.(uuid.UUID); ok {
			return u
		}
	}
	return uuid.Nil
}

// GetUserID 从 Context 中获取用户 ID (uint)
func GetUserID(c *gin.Context) uint {
	if v, exists := c.Get(ctxKeyUserID); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetAuthorityId 从 Context 中获取角色 ID
func GetAuthorityId(c *gin.Context) uint {
	if v, exists := c.Get(ctxKeyAuthorityID); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
