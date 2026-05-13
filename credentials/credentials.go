// Copyright JAMF Software, LLC

package credentials

import (
	"context"
	"crypto/tls"
	"net"
	"sync"

	grpccredentials "google.golang.org/grpc/credentials"
)

const (
	TokenFieldNameGRPC         = "token"
	AuthorizationFieldNameGRPC = "authorization"
	BearerTokenType            = "Bearer"
)

// TokenProvider resolves auth tokens for outgoing gRPC requests.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// TokenProviderFunc adapts a function to the TokenProvider interface.
type TokenProviderFunc func(ctx context.Context) (string, error)

// Token resolves an auth token for the provided request context.
func (f TokenProviderFunc) Token(ctx context.Context) (string, error) {
	return f(ctx)
}

// Config defines gRPC credential configuration.
type Config struct {
	TLSConfig                *tls.Config
	Token                    string
	TokenFieldName           string
	TokenType                string
	TokenProvider            TokenProvider
	RequireTransportSecurity bool
}

// Bundle defines gRPC credential interface.
type Bundle interface {
	grpccredentials.Bundle
	UpdateAuthToken(token string)
}

// NewBundle constructs a new gRPC credential bundle.
func NewBundle(cfg Config) Bundle {
	return &bundle{
		tc: newTransportCredential(cfg.TLSConfig),
		rc: newPerRPCCredential(cfg),
	}
}

// bundle implements "grpccredentials.Bundle" interface.
type bundle struct {
	tc *transportCredential
	rc *perRPCCredential
}

func (b *bundle) TransportCredentials() grpccredentials.TransportCredentials {
	return b.tc
}

func (b *bundle) PerRPCCredentials() grpccredentials.PerRPCCredentials {
	return b.rc
}

func (b *bundle) NewWithMode(mode string) (grpccredentials.Bundle, error) {
	// no-op
	return nil, nil
}

// transportCredential implements "grpccredentials.TransportCredentials" interface.
type transportCredential struct {
	gtc grpccredentials.TransportCredentials
}

func newTransportCredential(cfg *tls.Config) *transportCredential {
	return &transportCredential{
		gtc: grpccredentials.NewTLS(cfg),
	}
}

func (tc *transportCredential) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, grpccredentials.AuthInfo, error) {
	return tc.gtc.ClientHandshake(ctx, authority, rawConn)
}

func (tc *transportCredential) ServerHandshake(rawConn net.Conn) (net.Conn, grpccredentials.AuthInfo, error) {
	return tc.gtc.ServerHandshake(rawConn)
}

func (tc *transportCredential) Info() grpccredentials.ProtocolInfo {
	return tc.gtc.Info()
}

func (tc *transportCredential) Clone() grpccredentials.TransportCredentials {
	return &transportCredential{
		gtc: tc.gtc.Clone(),
	}
}

func (tc *transportCredential) OverrideServerName(serverNameOverride string) error {
	//nolint:staticcheck
	return tc.gtc.OverrideServerName(serverNameOverride)
}

// perRPCCredential implements "grpccredentials.PerRPCCredentials" interface.
type perRPCCredential struct {
	authToken                string
	authTokenMu              sync.RWMutex
	tokenFieldName           string
	tokenType                string
	tokenProvider            TokenProvider
	requireTransportSecurity bool
}

func newPerRPCCredential(cfg Config) *perRPCCredential {
	tokenFieldName := cfg.TokenFieldName
	if tokenFieldName == "" {
		tokenFieldName = TokenFieldNameGRPC
	}

	return &perRPCCredential{
		authToken:                cfg.Token,
		tokenFieldName:           tokenFieldName,
		tokenType:                cfg.TokenType,
		tokenProvider:            cfg.TokenProvider,
		requireTransportSecurity: cfg.RequireTransportSecurity,
	}
}

func (rc *perRPCCredential) RequireTransportSecurity() bool {
	return rc.requireTransportSecurity
}

func (rc *perRPCCredential) GetRequestMetadata(ctx context.Context, s ...string) (map[string]string, error) {
	authToken, err := rc.authTokenForContext(ctx)
	if err != nil {
		return nil, err
	}
	if authToken == "" {
		return nil, nil
	}

	return map[string]string{rc.tokenFieldName: rc.tokenValue(authToken)}, nil
}

func (rc *perRPCCredential) authTokenForContext(ctx context.Context) (string, error) {
	rc.authTokenMu.RLock()
	authToken := rc.authToken
	rc.authTokenMu.RUnlock()

	if rc.tokenProvider == nil {
		return authToken, nil
	}

	providerToken, err := rc.tokenProvider.Token(ctx)
	if err != nil {
		return "", err
	}
	if providerToken != "" {
		return providerToken, nil
	}

	return authToken, nil
}

func (rc *perRPCCredential) tokenValue(authToken string) string {
	if rc.tokenType == "" {
		return authToken
	}

	return rc.tokenType + " " + authToken
}

func (b *bundle) UpdateAuthToken(token string) {
	if b.rc == nil {
		return
	}
	b.rc.UpdateAuthToken(token)
}

func (rc *perRPCCredential) UpdateAuthToken(token string) {
	rc.authTokenMu.Lock()
	rc.authToken = token
	rc.authTokenMu.Unlock()
}
