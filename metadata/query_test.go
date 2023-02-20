package metadata

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Jguer/aur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cacheFilePath := dir + "/cache.json"

	// read test.json
	testBytes, err := os.ReadFile("test.json")
	require.NoError(t, err)

	client, err := New(
		WithCacheFilePath(cacheFilePath),
		WithHTTPClient(&MockHTTP{bytesToReturn: testBytes}),
		WithDebugLogger(func(s ...any) {
			t.Log(s...)
		}),
	)

	require.NoError(t, err)

	ctx := context.Background()

	type testcase struct {
		desc          string
		query         *aur.Query
		expectedNames []string
		wantPanic     bool
	}

	tests := []testcase{
		{
			desc: "single package",
			query: &aur.Query{
				By:       aur.Name,
				Needles:  []string{"jack-audio-tools-lv2"},
				Contains: false,
			},
			expectedNames: []string{"jack-audio-tools-lv2"},
		},
		{
			desc: "single package - known contains",
			query: &aur.Query{
				By:       aur.Name,
				Needles:  []string{"yay"},
				Contains: false,
			},
			expectedNames: []string{"yay"},
		},
		{
			desc: "starts-with",
			query: &aur.Query{
				By:       aur.Name,
				Needles:  []string{"jack-audio"},
				Contains: true,
			},
			expectedNames: []string{
				"jack-audio-tools-common",
				"jack-audio-tools-transport",
				"jack-audio-tools-lv2",
				"jack-audio-tools-carla",
				"jack-audio-tools-dbus"},
		},
		{
			desc: "contains",
			query: &aur.Query{
				By:       aur.Name,
				Needles:  []string{"tools"},
				Contains: true,
			},
			expectedNames: []string{
				"jack-audio-tools-common",
				"jack-audio-tools-transport",
				"jack-audio-tools-lv2",
				"jack-audio-tools-carla",
				"jack-audio-tools-dbus"},
		},
		{
			desc: "None", // Name + Provides
			query: &aur.Query{
				By:       aur.None,
				Needles:  []string{"yay"},
				Contains: true,
			},
			expectedNames: []string{
				"yay",
				"yay-bin",
				"yay-git",
			},
		},
		{
			desc: "Provides",
			query: &aur.Query{
				By:       aur.Provides,
				Needles:  []string{"yay"},
				Contains: false,
			},
			expectedNames: []string{
				"yay",
				"yay-bin",
				"yay-git",
			},
		},
		{
			desc: "None - Wireguard", // Name + Provides
			query: &aur.Query{
				By:       aur.None,
				Needles:  []string{"WIREGUARD-MODULE"},
				Contains: false,
			},
			expectedNames: []string{},
		},
		{
			desc: "Maintainer",
			query: &aur.Query{
				By:       aur.Maintainer,
				Needles:  []string{"jguer"},
				Contains: false,
			},
			expectedNames: []string{
				"yay",
				"yay-bin",
				"yay-git",
			},
		},
		{
			desc: "Submitter",
			query: &aur.Query{
				By:       aur.Submitter,
				Needles:  []string{"submitter"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "CheckDepends",
			query: &aur.Query{
				By:       aur.CheckDepends,
				Needles:  []string{"lv2lint"},
				Contains: false,
			},
			expectedNames: []string{"liquidsfz-git"},
		},
		{
			desc: "CheckDepends",
			query: &aur.Query{
				By:       aur.OptDepends,
				Needles:  []string{"libjack.so"},
				Contains: false,
			},
			expectedNames: []string{"liquidsfz-git"},
		},
		{
			desc: "Depends",
			query: &aur.Query{
				By:       aur.Depends,
				Needles:  []string{"kmod"},
				Contains: false,
			},
			expectedNames: []string{"linux-amd-git", "linux-ath-dfs"},
		},
		{
			desc: "MakeDepends",
			query: &aur.Query{
				By:       aur.MakeDepends,
				Needles:  []string{"pahole"},
				Contains: false,
			},
			expectedNames: []string{"linux-amd-git", "linux-ath-dfs"},
		},
		{
			desc: "NameDeps",
			query: &aur.Query{
				By:       aur.NameDesc,
				Needles:  []string{"Pre-compiled"},
				Contains: true,
			},
			expectedNames: []string{"yay-bin"},
		},
		{
			desc: "Conflicts",
			query: &aur.Query{
				By:       aur.Conflicts,
				Needles:  []string{"conflicts1"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "Replaces",
			query: &aur.Query{
				By:       aur.Replaces,
				Needles:  []string{"replaces1"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "Keywords",
			query: &aur.Query{
				By:       aur.Keywords,
				Needles:  []string{"keyword1"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "Groups",
			query: &aur.Query{
				By:       aur.Groups,
				Needles:  []string{"group1"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "CoMaintainers",
			query: &aur.Query{
				By:       aur.CoMaintainers,
				Needles:  []string{"comaintainer1"},
				Contains: false,
			},
			expectedNames: []string{"testpackage"},
		},
		{
			desc: "Panic",
			query: &aur.Query{
				By:       -10, // unsupported
				Needles:  []string{"prep"},
				Contains: true,
			},
			expectedNames: []string{""},
			wantPanic:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.wantPanic {
				assert.Panics(t, func() {
					client.Get(ctx, test.query)
				})
				return
			}

			pkgs, err := client.Get(ctx, test.query)
			require.NoError(t, err)

			var names []string
			for _, pkg := range pkgs {
				names = append(names, pkg.String())
			}

			assert.Len(t, pkgs, len(test.expectedNames))
			assert.ElementsMatch(t, test.expectedNames, names, fmt.Sprintf("%#v", names))
		})
	}
}
