package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	require.NotPanics(t, func() { doVersion(versionCmd, []string{}) })

}
