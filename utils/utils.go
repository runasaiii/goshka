package utils
import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)


const (
	jwtClaimType   = "token_type"
	jwtTypeAccess  = "access"
	jwtTypeRefresh = "refresh"
	bearerPrefix   = "Bearer "
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(plainPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func jwtSecretBytes() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return []byte(secret), nil
}

func issueToken(userID, role, tokenType string, ttl time.Duration) (string, error) {
	secret, err := jwtSecretBytes()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		jwtClaimType: tokenType,
		"exp":        time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GenerateAccessToken(userID, role string) (string, error) {
	return issueToken(userID, role, jwtTypeAccess, 15*time.Minute)
}

func GenerateRefreshToken(userID, role string) (string, error) {
	return issueToken(userID, role, jwtTypeRefresh, 7*24*time.Hour)
}

func GenerateJWT(userID, role string) (string, error) {
	return GenerateAccessToken(userID, role)
}

func stripBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization")
	}
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("invalid authorization scheme")
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix)), nil
}

func ParseBearerAccessToken(authHeader string) (userID, role string, err error) {
	tokenStr, err := stripBearer(authHeader)
	if err != nil {
		return "", "", err
	}
	return ParseAccessTokenString(tokenStr)
}

func ParseBearerAnyUserID(authHeader string) (userID string, err error) {
	tokenStr, err := stripBearer(authHeader)
	if err != nil {
		return "", err
	}
	uid, _, _, err := parseTokenClaims(tokenStr, "")
	return uid, err
}

func ParseAccessTokenString(tokenString string) (userID, role string, err error) {
	uid, r, typ, err := parseTokenClaims(tokenString, jwtTypeAccess)
	if err != nil {
		return "", "", err
	}
	if typ != jwtTypeAccess {
		return "", "", errors.New("expected access token")
	}
	return uid, r, nil
}

func ParseRefreshTokenString(tokenString string) (userID, role string, err error) {
	uid, r, typ, err := parseTokenClaims(tokenString, jwtTypeRefresh)
	if err != nil {
		return "", "", err
	}
	if typ != jwtTypeRefresh {
		return "", "", errors.New("expected refresh token")
	}
	return uid, r, nil
}

func parseTokenClaims(tokenString string, expectedType string) (userID, role, typ string, err error) {
	secret, err := jwtSecretBytes()
	if err != nil {
		return "", "", "", err
	}
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", "", "", err
	}
	uid, ok := claims["user_id"].(string)
	if !ok || uid == "" {
		return "", "", "", errors.New("invalid token: user_id")
	}
	r, _ := claims["role"].(string)
	tv, _ := claims[jwtClaimType].(string)
	if expectedType != "" && tv != expectedType {
		return "", "", tv, fmt.Errorf("invalid token type")
	}
	return uid, r, tv, nil
}
