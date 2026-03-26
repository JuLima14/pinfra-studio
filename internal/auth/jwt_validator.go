package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenMissing = errors.New("authorization token missing")
	ErrTokenInvalid = errors.New("authorization token invalid")
	ErrTokenExpired = errors.New("authorization token expired")
)

type Claims struct {
	jwt.RegisteredClaims
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}

type JWTValidator struct {
	domain string
	issuer string
	jwks   *JWKS
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func NewJWTValidator(domain string) *JWTValidator {
	return &JWTValidator{
		domain: domain,
		issuer: fmt.Sprintf("https://%s/", domain),
	}
}

func (v *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}
		return v.getPublicKey(kid)
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.Issuer != v.issuer {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	if v.jwks == nil {
		if err := v.fetchJWKS(); err != nil {
			return nil, err
		}
	}
	for _, key := range v.jwks.Keys {
		if key.Kid == kid {
			return v.jwkToRSAPublicKey(&key)
		}
	}
	return nil, fmt.Errorf("key with kid %s not found", kid)
}

func (v *JWTValidator) fetchJWKS() error {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", v.domain)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}
	v.jwks = &jwks
	return nil
}

func (v *JWTValidator) jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	if len(jwk.X5c) > 0 {
		certPEM := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", jwk.X5c[0])
		return jwt.ParseRSAPublicKeyFromPEM([]byte(certPEM))
	}
	return nil, errors.New("x5c not found in JWK")
}

func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrTokenMissing
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrTokenInvalid
	}
	return parts[1], nil
}
