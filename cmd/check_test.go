package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	require.PanicsWithError(t, "open no-such-file: no such file or directory", func() { doCheck(checkCmd, []string{"no-such-file"}) })
	require.NotPanics(t, func() { doCheck(checkCmd, []string{"testdata/node_exporter.yaml"}) })
	outfile = "testdata/converted.yaml"
	require.NotPanics(t, func() { doConvert(checkCmd, []string{"testdata/libvirt_scrape.txt"}) })
	defer os.Remove(outfile)
	require.NotPanics(t, func() { doCheck(checkCmd, []string{outfile}) })
	require.NotPanics(t, func() { doConvert(checkCmd, []string{"testdata/collectd_scrape.txt"}) })
	require.NotPanics(t, func() { doCheck(checkCmd, []string{outfile}) })

}
