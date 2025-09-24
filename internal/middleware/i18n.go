package middleware

import (
	"github.com/gin-gonic/gin"
)

// Package middleware -----------------------------
// @file        : i18n.go
// @author      : CIPFZ
// @time        : 2025/9/19 19:07
// @description :
// -------------------------------------------

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "en" // 默认英文
		}
		localizer := i18n.NewLocalizer(lang)
		c.Set("localizer", localizer)
		c.Next()
	}
}
