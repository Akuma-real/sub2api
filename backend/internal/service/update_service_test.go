package service

import "testing"

func TestCompareVersionsHandlesForkPrereleases(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{
			name:    "fork prerelease one patch ahead is newer than upstream stable",
			current: "0.1.127-akuma.1",
			latest:  "0.1.126",
			want:    1,
		},
		{
			name:    "fork prerelease on same patch is older than stable",
			current: "0.1.126-akuma.1",
			latest:  "0.1.126",
			want:    -1,
		},
		{
			name:    "v prefix is accepted",
			current: "v0.1.127-akuma.1",
			latest:  "v0.1.126",
			want:    1,
		},
		{
			name:    "numeric patch comparison remains semantic",
			current: "0.1.10",
			latest:  "0.1.9",
			want:    1,
		},
		{
			name:    "equal versions compare equal",
			current: "0.1.126",
			latest:  "v0.1.126",
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.current, tt.latest)
			if got != tt.want {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}
