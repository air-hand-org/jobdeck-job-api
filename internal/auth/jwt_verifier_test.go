package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	jwtauthv1 "github.com/air-hand-org/api-contracts/packages/jobdeck/auth/go/jsonschema/v1"
	"github.com/air-hand-org/jobdeck-job-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTVerifierVerifyFetchesAndCachesJWKSKey(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/jwks.json" {
			t.Errorf("request path = %q, want %q", r.URL.Path, "/.well-known/jwks.json")
		}
		if r.Header.Get("Accept") != "application/jwk-set+json" {
			t.Errorf("Accept header = %q, want %q", r.Header.Get("Accept"), "application/jwk-set+json")
		}
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/jwk-set+json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"alg": "EdDSA", "crv": "Ed25519", "kid": "key-1", "kty": "OKP", "use": "sig",
				"x": base64.RawURLEncoding.EncodeToString(publicKey),
			}},
		}); err != nil {
			t.Errorf("encode JWKS response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	verifier, err := NewJWTVerifier(server.URL)
	require.NoError(t, err)

	actorID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, accessTokenClaims{
		AccessTokenClaims: jwtauthv1.AccessTokenClaims{
			Iss: jwtauthv1.JobdeckAuth, Aud: []string{"jobdeck-job-api"}, Sub: actorID.String(),
			Iat: time.Now().Unix(), Exp: time.Now().Add(time.Minute).Unix(), ActorType: jwtauthv1.User,
		},
	})
	token.Header["kid"] = "key-1"
	rawToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	identity, err := verifier.Verify(context.Background(), rawToken)
	require.NoError(t, err)
	require.Equal(t, domain.AuthenticatedIdentity{ID: actorID, Type: domain.ActorTypeUser}, identity)

	_, err = verifier.Verify(context.Background(), rawToken)
	require.NoError(t, err)
	require.Equal(t, int32(1), requestCount.Load())
}

func TestJWTVerifierVerifyRejectsUnknownKeyID(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/jwk-set+json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"alg": "EdDSA", "crv": "Ed25519", "kid": "other-key", "kty": "OKP", "use": "sig",
				"x": base64.RawURLEncoding.EncodeToString(publicKey),
			}},
		}); err != nil {
			t.Errorf("encode JWKS response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	verifier, err := NewJWTVerifier(server.URL)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, accessTokenClaims{
		AccessTokenClaims: jwtauthv1.AccessTokenClaims{
			Iss: jwtauthv1.JobdeckAuth, Aud: []string{"jobdeck-job-api"}, Sub: uuid.NewString(),
			Iat: time.Now().Unix(), Exp: time.Now().Add(time.Minute).Unix(), ActorType: jwtauthv1.User,
		},
	})
	token.Header["kid"] = "missing-key"
	rawToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	_, err = verifier.Verify(context.Background(), rawToken)
	require.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestNewJWTVerifierRejectsInvalidAuthHTTPEndpoint(t *testing.T) {
	t.Parallel()

	_, err := NewJWTVerifier("not-a-url")
	require.EqualError(t, err, "invalid auth HTTP endpoint")
}

func TestNewJWTVerifierRejectsJWKSPath(t *testing.T) {
	t.Parallel()

	_, err := NewJWTVerifier("https://auth.example/.well-known/jwks.json")
	require.EqualError(t, err, "invalid auth HTTP endpoint")
}
