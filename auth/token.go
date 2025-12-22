package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	kindAccess  = "access"
	kindRefresh = "refresh"
)

type JWTTokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTTokenManager(secret string, accessTTL, refreshTTL time.Duration) *JWTTokenManager {
	return &JWTTokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *JWTTokenManager) signToken(userID string, exp time.Time, kind string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     exp.Unix(),
		"iat":     time.Now().Unix(),
		"kind":    kind, // Add this to differentiate tokens
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (j *JWTTokenManager) Generate(userID string) (*TokenPair, error) {
	now := time.Now()

	// Pass the kind during generation
	accessToken, err := j.signToken(userID, now.Add(j.accessTTL), kindAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.signToken(userID, now.Add(j.refreshTTL), kindRefresh)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *JWTTokenManager) verifyToken(tokenStr string, expectedKind string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secret, nil
	})

	if err != nil {
		// Log the actual internal error to see if it's "signature invalid" or "expired"
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}

	// Double check validity
	if !token.Valid {
		return "", ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrTokenInvalid
	}

	// 1. Verify 'kind'
	kind, _ := claims["kind"].(string)
	if kind != expectedKind {
		return "", ErrTokenInvalid
	}

	// 2. Verify 'user_id'
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", ErrTokenInvalid
	}

	return userID, nil
}

func (j *JWTTokenManager) VerifyAccess(token string) (string, error) {
	return j.verifyToken(token, kindAccess)
}

func (j *JWTTokenManager) VerifyRefresh(token string) (string, error) {
	return j.verifyToken(token, kindRefresh)
}
