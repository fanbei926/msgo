package token

import (
	"errors"
	"fanfan926.icu/msgo/v2"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

const JWTToken = "msgo_token"

type JWTHandler struct {
	Alg            string
	TimeOut        time.Duration
	RefreshTimeOut time.Duration
	TimeFunc       func() time.Time
	Key            []byte
	RefreshKey     string
	PrivateKey     string
	SendCookie     bool
	Authenticator  func(ctx *msgo.Context) (map[string]any, error)
	CookieName     string
	CookieMaxAge   int
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	Header         string
	AuthHandler    func(ctx *msgo.Context, err error)
}

type JWTResponse struct {
	Token        string
	RefreshToken string
}

// login
func (j *JWTHandler) LoginHandler(ctx *msgo.Context) (*JWTResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}

	if j.Alg == "" {
		j.Alg = "HS256"
	}

	// part A
	signingMethod := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(signingMethod)
	// part B
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}

	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}

	expire := j.TimeFunc().Add(j.TimeOut)
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFunc().Unix()
	var tokenString string
	var tokenErr error

	// part C
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}

	if tokenErr != nil {
		return nil, tokenErr
	}

	jr := &JWTResponse{
		Token: tokenString,
	}

	//refresh token
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	// send cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFunc().Unix())
		}
		ctx.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

func (j *JWTHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JWTHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFunc().Add(j.RefreshTimeOut).Unix()

	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}

	if tokenErr != nil {
		return "", tokenErr
	}

	return tokenString, nil
}

func (j *JWTHandler) LogoutHandler(ctx *msgo.Context) error {
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		return nil
	}
	return nil
}

func (j *JWTHandler) RefreshHandler(ctx *msgo.Context) (*JWTResponse, error) {
	rToken, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}

	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// parse token
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if j.usingPublicKeyAlgo() {
			return j.PrivateKey, nil
		} else {
			return j.Key, nil
		}
	})

	if err != nil {
		return nil, err
	}

	// part B
	claims := t.Claims.(jwt.MapClaims)
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}

	expire := j.TimeFunc().Add(j.TimeOut)
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFunc().Unix()
	var tokenString string
	var tokenErr error

	// part C
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = t.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(j.Key)
	}

	if tokenErr != nil {
		return nil, tokenErr
	}

	jr := &JWTResponse{
		Token: tokenString,
	}

	//refresh token
	refreshToken, err := j.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	// send cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFunc().Unix())
		}
		ctx.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

// jwt middlewares
func (j *JWTHandler) AuthInterceptor(next msgo.HandleFunc) msgo.HandleFunc {
	return func(ctx *msgo.Context) {
		if j.Header == "" {
			j.Header = "Authorization"
		}
		token := ctx.R.Header.Get(j.Header)
		if token == "" {
			if j.SendCookie {
				cookie, err := ctx.R.Cookie(j.CookieName)
				if err != nil {
					j.AuthErrorHandler(ctx, err)
					return
				}
				token = cookie.String()
			}

		}

		if token == "" {
			j.AuthErrorHandler(ctx, errors.New("token xxxxx"))
			return
		}

		// parse token
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if j.usingPublicKeyAlgo() {
				return []byte(j.PrivateKey), nil
			} else {
				return []byte(j.Key), nil
			}
		})

		if err != nil {
			j.AuthErrorHandler(ctx, err)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		ctx.Set("jwt_claims", claims)
		next(ctx)
	}
}

func (j *JWTHandler) AuthErrorHandler(ctx *msgo.Context, err error) {
	if j.AuthHandler == nil {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	} else {
		j.AuthHandler(ctx, err)
	}
}
