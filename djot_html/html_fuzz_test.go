package html_writer

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func seedFuzz(f *testing.F) {
	dir, err := os.ReadDir(examplesDir)
	require.Nil(f, err)
	for _, entry := range dir {
		name := entry.Name()
		example, ok := strings.CutSuffix(name, ".html")
		if !ok {
			continue
		}
		djotExample, err := os.ReadFile(path.Join(examplesDir, fmt.Sprintf("%v.djot", example)))
		require.Nil(f, err)
		f.Add(string(djotExample))
	}
}

func FuzzDjotE2E(f *testing.F) {
	seedFuzz(f)
	f.Fuzz(func(t *testing.T, input string) { _ = printDjot(input) })
}
