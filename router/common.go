package router

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

const (
	defaultSize int = 10
	defaultPage int = 1
)

type responseWithPaging struct {
	Records     interface{}            `json:"records"`
	PageSummary map[string]interface{} `json:"page_summary"`
}

func withPaging(result interface{}, total int64, page, size int) responseWithPaging {
	offset := (page - 1) * size

	var hasNext bool
	if offset+size < int(total) {
		hasNext = true
	}

	return responseWithPaging{
		Records: result,
		PageSummary: map[string]interface{}{
			"size":    size,
			"page":    page,
			"hasNext": hasNext,
			"total":   total,
		},
	}
}

func getOffset(page, size int) int {
	return (page - 1) * size
}

type jwtClaims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (j *jwtClaims) IsSuperAdmin() bool {
	return j.Role == "notblessy"
}

func (j *jwtClaims) IsUser() bool {
	return j.Role == "user"
}

type JWTMiddleware struct{}

func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{}
}

func (m *JWTMiddleware) ValidateJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, response{
				Message: "authorization token is required",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.JSON(http.StatusUnauthorized, response{
				Message: "token is malformed",
			})
		}

		// Call gRPC to validate the token
		user, err := validateToken(token)
		if err != nil || user.ID == "" {
			logrus.Error(err)
			return c.JSON(http.StatusUnauthorized, response{
				Message: "cannot validate token: " + err.Error(),
			})
		}

		c.Set("user", user)

		return next(c)
	}
}

func validateToken(tokenString string) (jwtClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return jwtClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return jwtClaims{}, errors.New("invalid token claims")
	}

	uid, ok := claims["id"].(string)
	if !ok {
		return jwtClaims{}, errors.New("user id not found in claims")
	}

	name, ok := claims["name"].(string)
	if !ok {
		return jwtClaims{}, errors.New("name not found in claims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return jwtClaims{}, errors.New("roleId not found in claims")
	}

	return jwtClaims{
		ID:   uid,
		Name: name,
		Role: role,
	}, nil
}

func signJwtToken(id, name string, role string) (string, error) {
	claims := &jwtClaims{
		ID:   id,
		Name: name,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(24*7))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return t, nil
}

func authSession(c echo.Context) (jwtClaims, error) {
	u := c.Get("user")
	if u == nil {
		return jwtClaims{}, errors.New("missing session")
	}

	user, ok := u.(jwtClaims)
	if !ok {
		return jwtClaims{}, errors.New("invalid session")
	}

	return user, nil
}
