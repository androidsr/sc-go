package sjwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/androidsr/sc-go/model"
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
	data["notBefore"] = time.Now().Unix()
	data["expiresAt"] = time.Now().Add(time.Duration(config.Expire) * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenString, err = token.SignedString([]byte(config.SecretKey))
	return
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func SetWebToken(c *gin.Context, tokenStr string) {
	switch config.StoreType {
	case 1:
		c.Header(config.TokenName, "Bearer "+tokenStr)
	case 2:
		c.SetCookie(config.TokenName, "Bearer "+tokenStr, 60*config.Expire, "", "", false, true)
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

		parts := strings.SplitN(tokenStr, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "JSON Web Token 格式不正确;示例【Bearer xxx.xxx.xxx】"))
			c.Abort()
			return
		}

		mc, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "无效的Token"))
			c.Abort()
			return
		}
		// 检查JWT令牌是否快要过期
		expiresAt, ok := mc["expiresAt"].(int64)
		if !ok {
			c.JSON(http.StatusUnauthorized, model.NewFail(401, "无效的Token"))
			c.Abort()
		}
		if (expiresAt-time.Now().Unix())/60 < int64(config.Expire)*60/3 {
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
