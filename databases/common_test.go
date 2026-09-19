package databases

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionURLEscapesCredentials(t *testing.T) {
	dsn := connectionURL("2001:db8::1", 5432, "o'hara@example.com", "p'a ss?&/%")

	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsed.Scheme)
	require.Equal(t, "[2001:db8::1]:5432", parsed.Host)
	require.Equal(t, "o'hara@example.com", parsed.User.Username())
	password, ok := parsed.User.Password()
	require.True(t, ok)
	require.Equal(t, "p'a ss?&/%", password)
	require.Equal(t, "disable", parsed.Query().Get("sslmode"))
}
