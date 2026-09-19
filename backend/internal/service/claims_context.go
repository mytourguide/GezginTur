package service

import "context"

// ctxKey, context'te claims saklamak icin kullanilan ozel tip.
type ctxKey string

const claimsCtxKey ctxKey = "jwt-claims"

// ContextWithClaims, request context'ine JWT claim'lerini yazar.
func ContextWithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey, c)
}

// ClaimsFrom, context'ten JWT claim'lerini okur (middleware yazar).
func ClaimsFrom(ctx context.Context) *Claims {
	if c, ok := ctx.Value(claimsCtxKey).(*Claims); ok {
		return c
	}
	return &Claims{} // bos claims: kimliksiz isteklerde guvenli sifir deger
}
