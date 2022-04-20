package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	require.PanicsWithError(t, "open no-such-file: no such file or directory", func() { doConvert(convertCmd, []string{"no-such-file"}) })
	outfile = "/dev/null"
	require.NotPanics(t, func() { doConvert(convertCmd, []string{"testdata/libvirt_scrape.txt"}) })
	require.NotPanics(t, func() { doConvert(convertCmd, []string{"testdata/collectd_scrape.txt"}) })
}
