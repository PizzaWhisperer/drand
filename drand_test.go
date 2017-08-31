package main

import (
	"os"
	"sync"
	"testing"

	"github.com/nikkolasg/slog"
	"github.com/stretchr/testify/require"
)

func TestDrandDKG(t *testing.T) {
	slog.Level = slog.LevelDebug
	n := 4
	_, drands := BatchDrands(n)
	defer CloseAllDrands(drands)

	shareFile := defaultShareFile()
	defer os.Remove(shareFile)

	var wg sync.WaitGroup
	wg.Add(n - 1)
	for _, drand := range drands[1:] {
		go func(d *Drand) {
			err := d.RunDKG(shareFile)
			require.Nil(t, err)
			wg.Done()
		}(drand)
	}
	root := drands[0]
	err := root.StartDKG(shareFile)
	require.Nil(t, err)
	wg.Wait()
}