package aur

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBy_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		by   By
		want string
	}{
		{
			name: "name", by: Name, want: "name",
		},
		{
			name: "namedesc", by: NameDesc, want: "name-desc",
		},
		{
			name: "maintainer", by: Maintainer, want: "maintainer",
		},
		{
			name: "submitter", by: Submitter, want: "submitter",
		},
		{
			name: "depends", by: Depends, want: "depends",
		},
		{
			name: "makedepends", by: MakeDepends, want: "makedepends",
		},
		{
			name: "optdepends", by: OptDepends, want: "optdepends",
		},
		{
			name: "optdepends", by: OptDepends, want: "optdepends",
		},
		{
			name: "checkdepends", by: CheckDepends, want: "checkdepends",
		},
		{
			name: "provides", by: Provides, want: "provides",
		},
		{
			name: "replaces", by: Replaces, want: "replaces",
		},
		{
			name: "conflicts", by: Conflicts, want: "conflicts",
		},
		{
			name: "keywords", by: Keywords, want: "keywords",
		},
		{
			name: "groups", by: Groups, want: "groups",
		},
		{
			name: "comaintainers", by: CoMaintainers, want: "comaintainers",
		},
		{
			name: "default", by: None, want: "",
		},
		{
			name: "panic", by: 23, want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.by == 23 {
				assert.Panics(t, func() { _ = tt.by.String() })
			} else {
				got := tt.by.String()
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
