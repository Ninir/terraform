package archive

import (
	"testing"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestResolveMatches(t *testing.T) {
	inclusions := []string{}
	exclusions := []string{}

	archiver := NewZipArchiver(zipfilepath)
	if err := archiver.ArchiveContent([]byte("This is some content"), "content.txt"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	resolveMatches(inclusions, exclusions)
}
