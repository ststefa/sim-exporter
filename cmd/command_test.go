package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	require.PanicsWithError(t, "open no-such-file: no such file or directory", func() { doConvert(convertCmd, []string{"no-such-file"}) })
	outfile = "/dev/null"
	require.NotPanics(t, func() { doConvert(convertCmd, []string{"testdata/libvirt_scrape.txt"}) })
}

func TestVersion(t *testing.T) {
	require.NotPanics(t, func() { doVersion(versionCmd, []string{}) })

}
func TestCheck(t *testing.T) {
	require.PanicsWithError(t, "open no-such-file: no such file or directory", func() { doCheck(convertCmd, []string{"no-such-file"}) })
	require.NotPanics(t, func() { doCheck(convertCmd, []string{"testdata/node_exporter.yaml"}) })
	outfile = "testdata/converted.yaml"
	require.NotPanics(t, func() { doConvert(convertCmd, []string{"testdata/libvirt_scrape.txt"}) })
	defer os.Remove(outfile)
	require.NotPanics(t, func() { doCheck(convertCmd, []string{outfile}) })
	require.NotPanics(t, func() { doConvert(convertCmd, []string{"testdata/collectd_scrape.txt"}) })
	require.NotPanics(t, func() { doCheck(convertCmd, []string{outfile}) })

}
