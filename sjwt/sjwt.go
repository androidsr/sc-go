package sjwt

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/androidsr/sc-go/model"
	"github.com/androidsr/sc-go/sc"
	"github.com/androidsr/sc-go/syaml"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	Authorization = "Authorization"
)

var (
	config *syaml.WebTokenInfo
)

func New(cfg *syaml.WebTokenInfo) {
	config = cfg
}

// 生成 Token
func GenerateToken(data jwt.MapClaims) (tokenString string, err error) {
	data["notBefore"] = sc.FormatDateTimeString(time.Now())
	data["expiresAt"] = sc.FormatDateTimeString(time.Now().Add(time.Duration(config.Expire) * time.Minute))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenString, err = token.SignedString([]byte(config.SecretKey))
	return
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func SetWebToken(c *gin.Context, tokenStr string) {
	switch config.StoreType {
	case 1:
		c.Header(config.TokenName, tokenStr)
	case 2:
		c.SetCookie(config.TokenName, tokenStr, 60*config.Expire, "", "", false, true)
	}
}

// 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		url := c.Request.URL.Path
		for _, v := range config.WhiteList {
			if v == url {
				c.Next()
				return
			}
		}
		var tokenStr string
		switch config.StoreType {
		case 1:
			tokenStr = c.Request.Header.Get(config.TokenName)
		case 2:
			tokenStr, _ = c.Cookie(config.TokenName)
		case 3:
			tokenStr = c.Param(config.TokenName)
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "请先登录后重试！"))
			c.Abort()
			return
		}
		mc, err := ParseToken(tokenStr)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "无效的Token"))
			c.Abort()
			return
		}
		// 检查JWT令牌是否快要过期
		e, ok := mc["expiresAt"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "无效的Token"))
			c.Abort()
		}

		expiresAt := sc.ParseDateTime(e)
		if expiresAt.Sub(time.Now().Local()).Minutes() <= float64(config.Expire/3) {
			// 如果快要过期，则刷新JWT令牌
			tokenStr, _ := GenerateToken(mc)
			switch config.StoreType {
			case 1:
				c.Writer.Header().Set(config.TokenName, tokenStr)
			case 2:
				c.SetCookie(config.TokenName, tokenStr, config.Expire*60, "/", "localhost", false, true)
			case 3:
				return
			}
		}
		for k, v := range mc {
			c.Set(k, v)
		}
		c.Next()
	}
}
