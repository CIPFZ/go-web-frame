package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/extlogin"
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/pkg/response"
)

// 外部登录（见 core/extlogin）：
//  1. GET  /user/ext-login/:provider   设置 state，把浏览器送到提供方的登录页
//  2. GET  /user/ext-callback/:provider 提供方带回来；核对 state、换回身份、关联本地账号，带一次性票据回登录页
//  3. POST /user/ext-ticket             登录页用票据换登录凭证

const (
	extStateCookie = "ext_login_state"
	extStateTTL    = 10 * time.Minute
)

// ExternalLogin 外部登录第一步：把浏览器送到提供方的登录页。
func (u *UserApi) ExternalLogin(c *gin.Context) {
	key := c.Param("provider")
	redirect := safeRedirect(c.Query("redirect"))
	provider, _, ok := extlogin.Get(key)
	if !ok {
		c.Redirect(http.StatusFound, extLoginPage(url.Values{"ext_error": {"unknown"}}, redirect))
		return
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		c.Redirect(http.StatusFound, extLoginPage(url.Values{"ext_error": {"failed"}}, redirect))
		return
	}
	state := hex.EncodeToString(raw)
	// state 同时放在回调地址和只给回调用的 Cookie 里：回调时两者一致才继续（防止别人伪造回调把自己的账号登录到你的浏览器上）
	q := url.Values{"state": {state}}
	if redirect != "" {
		q.Set("redirect", redirect)
	}
	callbackPath := u.svcCtx.Config.System.RouterPrefix + "/user/ext-callback/" + key
	c.SetSameSite(http.SameSiteLaxMode)
	secure := strings.HasPrefix(externalOrigin(c), "https://")
	c.SetCookie(extStateCookie, state, int(extStateTTL.Seconds()), callbackPath, "", secure, true)
	target, err := provider.LoginURL(c.Request.Context(), externalOrigin(c)+callbackPath+"?"+q.Encode())
	if err != nil {
		logger.GetLogger(c).Warn("ext_login_url_failed", zap.String("provider", key), zap.Error(err))
		c.Redirect(http.StatusFound, extLoginPage(url.Values{"ext_error": {"failed"}}, redirect))
		return
	}
	c.Redirect(http.StatusFound, target)
}

// ExternalCallback 提供方登录后的回调：核对 state，用回调参数换回身份，找到或关联本地用户，带一次性票据回登录页。
func (u *UserApi) ExternalCallback(c *gin.Context) {
	log := logger.GetLogger(c)
	key := c.Param("provider")
	redirect := safeRedirect(c.Query("redirect"))
	fail := func(reason string) {
		c.Redirect(http.StatusFound, extLoginPage(url.Values{"ext_error": {reason}}, redirect))
	}
	provider, opt, ok := extlogin.Get(key)
	if !ok {
		fail("unknown")
		return
	}
	cookie, _ := c.Cookie(extStateCookie)
	c.SetCookie(extStateCookie, "", -1, u.svcCtx.Config.System.RouterPrefix+"/user/ext-callback/"+key, "", false, true)
	state := c.Query("state")
	if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie)) != 1 {
		fail("state")
		return
	}
	identity, err := provider.Exchange(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		log.Warn("ext_login_exchange_failed", zap.String("provider", key), zap.Error(err))
		fail("failed")
		return
	}
	ticket, err := u.userService.ExternalLogin(c.Request.Context(), key, opt, identity)
	if err != nil {
		reason := "failed"
		switch {
		case errors.Is(err, service.ErrUserDisabled):
			reason = "user_disabled"
		case errors.Is(err, service.ErrExternalNotLinked), errors.Is(err, service.ErrExternalUsernameTaken):
			reason = "not_linked"
		default:
			log.Error("ext_login_failed", zap.String("provider", key), zap.Error(err))
		}
		fail(reason)
		return
	}
	c.Redirect(http.StatusFound, extLoginPage(url.Values{"ext_ticket": {ticket}}, redirect))
}

// ExternalTicket 用回调带回的一次性票据换登录凭证。
func (u *UserApi) ExternalTicket(c *gin.Context) {
	var req dto.LoginTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	resp, err := u.userService.RedeemLoginTicket(c.Request.Context(), req.Ticket)
	if err != nil {
		response.FailWithMessage("invalid username or password", c)
		return
	}
	u.setTokenHelper(c, resp.Token, int((resp.ExpiresAt/1000)-time.Now().Unix()))
	response.OkWithDetailed(resp, "login successful", c)
}

// externalOrigin 浏览器看到的站点地址（经前端 nginx 转发，协议取 X-Forwarded-Proto）。
func externalOrigin(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme != "http" && scheme != "https" {
		scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}
	return scheme + "://" + c.Request.Host
}

// extLoginPage 回到前端登录页（hash 路由），带上结果与登录后要去的页面。
func extLoginPage(q url.Values, redirect string) string {
	if redirect != "" {
		q.Set("redirect", redirect)
	}
	return "/#/user/login?" + q.Encode()
}

// safeRedirect 只接受站内路径，防止被当成跳转到外站的跳板。
func safeRedirect(r string) string {
	if !strings.HasPrefix(r, "/") || strings.HasPrefix(r, "//") || strings.Contains(r, "\\") {
		return ""
	}
	return r
}
