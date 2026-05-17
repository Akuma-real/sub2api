package service

import (
	"context"
	"testing"
	"time"
)

type noopUpdateCache struct{}

func (noopUpdateCache) GetUpdateInfo(context.Context) (string, error) {
	return "", context.Canceled
}

func (noopUpdateCache) SetUpdateInfo(context.Context, string, time.Duration) error {
	return nil
}

type recordingReleaseClient struct {
	repo string
}

func (c *recordingReleaseClient) FetchLatestRelease(_ context.Context, repo string) (*GitHubRelease, error) {
	c.repo = repo
	return &GitHubRelease{
		TagName: "v0.1.127-akuma.4",
		Name:    "Sub2API 0.1.127-akuma.4",
	}, nil
}

func (*recordingReleaseClient) DownloadFile(context.Context, string, string, int64) error {
	return nil
}

func (*recordingReleaseClient) FetchChecksumFile(context.Context, string) ([]byte, error) {
	return nil, nil
}

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

func TestUpdateServiceUsesConfiguredGitHubRepo(t *testing.T) {
	client := &recordingReleaseClient{}
	svc := NewUpdateService(noopUpdateCache{}, client, "0.1.127-akuma.3", "release", "example/sub2api")

	info, err := svc.CheckUpdate(context.Background(), true)
	if err != nil {
		t.Fatalf("CheckUpdate() returned error: %v", err)
	}
	if client.repo != "example/sub2api" {
		t.Fatalf("repo = %q, want %q", client.repo, "example/sub2api")
	}
	if !info.HasUpdate {
		t.Fatalf("HasUpdate = false, want true")
	}
}

func TestUpdateServiceDefaultsToForkGitHubRepo(t *testing.T) {
	client := &recordingReleaseClient{}
	svc := NewUpdateService(noopUpdateCache{}, client, "0.1.127-akuma.3", "release", "")

	_, err := svc.CheckUpdate(context.Background(), true)
	if err != nil {
		t.Fatalf("CheckUpdate() returned error: %v", err)
	}
	if client.repo != defaultGitHubRepo {
		t.Fatalf("repo = %q, want %q", client.repo, defaultGitHubRepo)
	}
}
