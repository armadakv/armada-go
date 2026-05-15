// Copyright JAMF Software, LLC

package credentials

import (
	"context"
	"errors"
	"testing"
)

func TestBundleUsesDefaultTokenMetadata(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(Config{Token: "static-token"})
	metadata, err := bundle.PerRPCCredentials().GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got := metadata[TokenFieldNameGRPC]; got != "static-token" {
		t.Fatalf("GetRequestMetadata() token = %q, want %q", got, "static-token")
	}

	if bundle.PerRPCCredentials().RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity() = true, want false")
	}
}

func TestBundleUsesBearerProviderMetadata(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(Config{
		Token:                    "fallback-token",
		TokenFieldName:           AuthorizationFieldNameGRPC,
		TokenType:                BearerTokenType,
		TokenProvider:            TokenProviderFunc(func(context.Context) (string, error) { return "provider-token", nil }),
		RequireTransportSecurity: true,
	})

	metadata, err := bundle.PerRPCCredentials().GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got := metadata[AuthorizationFieldNameGRPC]; got != "Bearer provider-token" {
		t.Fatalf("GetRequestMetadata() authorization = %q, want %q", got, "Bearer provider-token")
	}

	if !bundle.PerRPCCredentials().RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity() = false, want true")
	}
}

func TestBundleFallsBackToStaticTokenWhenProviderReturnsEmpty(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(Config{
		Token:          "fallback-token",
		TokenFieldName: AuthorizationFieldNameGRPC,
		TokenType:      BearerTokenType,
		TokenProvider:  TokenProviderFunc(func(context.Context) (string, error) { return "", nil }),
	})

	metadata, err := bundle.PerRPCCredentials().GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got := metadata[AuthorizationFieldNameGRPC]; got != "Bearer fallback-token" {
		t.Fatalf("GetRequestMetadata() authorization = %q, want %q", got, "Bearer fallback-token")
	}
}

func TestBundleReturnsProviderErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	bundle := NewBundle(Config{
		TokenProvider: TokenProviderFunc(func(context.Context) (string, error) {
			return "", wantErr
		}),
	})

	_, err := bundle.PerRPCCredentials().GetRequestMetadata(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetRequestMetadata() error = %v, want %v", err, wantErr)
	}
}

func TestBundleUpdateAuthTokenOverridesStaticToken(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(Config{
		TokenFieldName: AuthorizationFieldNameGRPC,
		TokenType:      BearerTokenType,
	})
	bundle.UpdateAuthToken("updated-token")

	metadata, err := bundle.PerRPCCredentials().GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got := metadata[AuthorizationFieldNameGRPC]; got != "Bearer updated-token" {
		t.Fatalf("GetRequestMetadata() authorization = %q, want %q", got, "Bearer updated-token")
	}
}
