package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/air-hand-org/jobdeck-job-api/internal/app"
	"github.com/air-hand-org/jobdeck-job-api/internal/domain"

	jwtauthv1 "github.com/air-hand-org/api-contracts/packages/jobdeck/auth/go/jsonschema/v1"
	jwksclient "github.com/air-hand-org/api-contracts/packages/jobdeck/auth/go/openapi/v1/jwks/client"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// accessTokenClaims adapts the shared access-token contract for the JWT library.
type accessTokenClaims struct {
	jwtauthv1.AccessTokenClaims
}

func (c accessTokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

func (c accessTokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.Iat == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Iat, 0)), nil
}

func (accessTokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c accessTokenClaims) GetIssuer() (string, error) {
	return string(c.Iss), nil
}

func (c accessTokenClaims) GetSubject() (string, error) {
	return c.Sub, nil
}

func (c accessTokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(c.Aud), nil
}

var _ app.TokenVerifier = (*JWTVerifier)(nil)

// JWTVerifier validates actor JWTs.
type JWTVerifier struct {
	client   *jwksclient.APIClient
	issuer   string
	audience string

	mu   sync.RWMutex
	keys map[string]ed25519.PublicKey
}

// Verify validates a token and returns its identity.
func (v *JWTVerifier) Verify(ctx context.Context, rawToken string) (domain.AuthenticatedIdentity, error) {
	claims := accessTokenClaims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		&claims,
		func(token *jwt.Token) (any, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok || kid == "" {
				return nil, errors.New("JWT kid is required")
			}
			return v.publicKey(ctx, kid)
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(30*time.Second),
	)

	if err != nil || !token.Valid {
		return domain.AuthenticatedIdentity{}, domain.ErrTokenInvalid
	}

	actorID, err := uuid.Parse(claims.Sub)
	if err != nil {
		return domain.AuthenticatedIdentity{}, domain.ErrTokenInvalid
	}

	return domain.AuthenticatedIdentity{
		ID:   actorID,
		Type: domain.ActorType(claims.ActorType),
	}, nil
}

func decodePublicKey(encoded string) (ed25519.PublicKey, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, errors.New("invalid Ed25519 public key")
	}
	return ed25519.PublicKey(decoded), nil
}

func (v *JWTVerifier) publicKey(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	v.mu.RUnlock()
	if ok {
		return key, nil
	}

	keySet, _, err := v.client.DefaultAPI.GetJSONWebKeySet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("get JWKS: %w", err)
	}

	keys := make(map[string]ed25519.PublicKey, len(keySet.Keys))
	for _, jwk := range keySet.Keys {
		if jwk.Alg != jwt.SigningMethodEdDSA.Alg() || jwk.Crv != "Ed25519" || jwk.Kty != "OKP" || jwk.Use != "sig" || jwk.Kid == "" {
			continue
		}

		publicKey, err := decodePublicKey(jwk.X)
		if err != nil {
			continue
		}
		keys[jwk.Kid] = publicKey
	}

	v.mu.Lock()
	v.keys = keys
	key, ok = v.keys[kid]
	v.mu.Unlock()
	if !ok {
		return nil, errors.New("JWT signing key not found")
	}
	return key, nil
}

// NewJWTVerifier creates a verifier that obtains public signing keys from a JWKS endpoint.
func NewJWTVerifier(jwksURL string) (*JWTVerifier, error) {
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(jwksURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("invalid JWKS URL")
	}

	configuration := jwksclient.NewConfiguration()
	configuration.Servers = jwksclient.ServerConfigurations{{URL: strings.TrimRight(parsedURL.String(), "/")}}
	configuration.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	return &JWTVerifier{
		client:   jwksclient.NewAPIClient(configuration),
		issuer:   string(jwtauthv1.JobdeckAuth),
		audience: "jobdeck-job-api",
		keys:     make(map[string]ed25519.PublicKey),
	}, nil
}
