package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cloud-mill/webrtc-signalling/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type JwtClaim struct {
	CustomClaims UserCustomClaim
	jwt.RegisteredClaims
}

type UserCustomClaim struct {
	AccountID uuid.UUID
	Username  string
	Email     string
}

type AuthConfig struct {
	JwtCookieName  string
	CsrfCookieName string
	CsrfHeaderName string
}

func ConvertToByteSecretKey(secretKey interface{}) ([]byte, error) {
	switch sk := secretKey.(type) {
	case string:
		return []byte(sk), nil
	case []byte:
		return sk, nil
	default:
		return nil, fmt.Errorf("invalid secret key type: %T", secretKey)
	}
}

func validateJWTAndGetClaims(
	tokenStr string,
	claims *JwtClaim,
	secretKey interface{},
) (jwt.Claims, error) {
	byteSecretKey, err := ConvertToByteSecretKey(secretKey)
	if err != nil {
		return nil, err
	}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return byteSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, errors.New("invalid token")
	}

	return tkn.Claims, nil
}

func AuthMiddleware(
	next http.Handler,
	secretKey interface{},
	authConfig AuthConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JWT cookie validation
		cookie, err := r.Cookie(authConfig.JwtCookieName)
		if err != nil {
			handleAuthError(w, err, "missing or invalid JWT cookie")
			return
		}

		tokenStr := cookie.Value
		jwtClaims := JwtClaim{}
		_, err = validateJWTAndGetClaims(tokenStr, &jwtClaims, secretKey)
		if err != nil {
			logger.Logger.Warn("invalid JWT token",
				zap.String("reason", err.Error()),
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// CSRF cookie validation
		csrfCookie, err := r.Cookie(authConfig.CsrfCookieName)
		if err != nil {
			handleAuthError(w, err, "missing or invalid CSRF cookie")
			return
		}

		csrfHeader := r.Header.Get(authConfig.CsrfHeaderName)
		if csrfHeader == "" || csrfHeader != csrfCookie.Value {
			logger.Logger.Warn("CSRF mismatch",
				zap.String("expected", csrfCookie.Value),
				zap.String("got", csrfHeader),
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleAuthError(w http.ResponseWriter, err error, msg string) {
	if errors.Is(err, http.ErrNoCookie) {
		logger.Logger.Warn("unauthorised access", zap.String("reason", msg))
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		logger.Logger.Error("auth middleware error",
			zap.String("reason", msg),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
	}
}
