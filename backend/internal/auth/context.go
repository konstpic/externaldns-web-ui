package auth

import "context"

type ctxKey int

const claimsKey ctxKey = 1

func withClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func claimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}

func UserFromContext(ctx context.Context) *Claims {
	return claimsFromContext(ctx)
}
