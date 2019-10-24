package k8sresolver

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTarget(t *testing.T) {
	tgt, err := parseTarget("ns-a", "svc-a:http", "default")
	require.NoError(t, err)
	require.Equal(t, "ns-a", tgt.Namespace)
	require.Equal(t, "http", tgt.Port)
	require.True(t, tgt.PortIsName)
	require.False(t, tgt.PortIsFirst)

	tgt, err = parseTarget("", "svc-a:http", "default")
	require.NoError(t, err)
	require.Equal(t, "default", tgt.Namespace)
	require.Equal(t, "http", tgt.Port)
	require.True(t, tgt.PortIsName)
	require.False(t, tgt.PortIsFirst)

	tgt, err = parseTarget("", "svc-a.ns-b:http", "default")
	require.NoError(t, err)
	require.Equal(t, "ns-b", tgt.Namespace)
	require.Equal(t, "http", tgt.Port)
	require.True(t, tgt.PortIsName)
	require.False(t, tgt.PortIsFirst)

	tgt, err = parseTarget("", "svc-a.ns-b:80", "default")
	require.NoError(t, err)
	require.Equal(t, "ns-b", tgt.Namespace)
	require.Equal(t, "80", tgt.Port)
	require.False(t, tgt.PortIsName)
	require.False(t, tgt.PortIsFirst)
}
