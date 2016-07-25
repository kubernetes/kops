package storage

import (
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/testutil"
)

func TestLinkedBlobStoreCreateWithMountFrom(t *testing.T) {
	fooRepoName, _ := reference.ParseNamed("nm/foo")
	fooEnv := newManifestStoreTestEnv(t, fooRepoName, "thetag")
	ctx := context.Background()

	// Build up some test layers and add them to the manifest, saving the
	// readseekers for upload later.
	testLayers := map[digest.Digest]io.ReadSeeker{}
	for i := 0; i < 2; i++ {
		rs, ds, err := testutil.CreateRandomTarFile()
		if err != nil {
			t.Fatalf("unexpected error generating test layer file")
		}
		dgst := digest.Digest(ds)

		testLayers[digest.Digest(dgst)] = rs
	}

	// upload the layers to foo/bar
	for dgst, rs := range testLayers {
		wr, err := fooEnv.repository.Blobs(fooEnv.ctx).Create(fooEnv.ctx)
		if err != nil {
			t.Fatalf("unexpected error creating test upload: %v", err)
		}

		if _, err := io.Copy(wr, rs); err != nil {
			t.Fatalf("unexpected error copying to upload: %v", err)
		}

		if _, err := wr.Commit(fooEnv.ctx, distribution.Descriptor{Digest: dgst}); err != nil {
			t.Fatalf("unexpected error finishing upload: %v", err)
		}
	}

	// create another repository nm/bar
	barRepoName, _ := reference.ParseNamed("nm/bar")

	barRepo, err := fooEnv.registry.Repository(ctx, barRepoName)
	if err != nil {
		t.Fatalf("unexpected error getting repo: %v", err)
	}

	// cross-repo mount the test layers into a nm/bar
	for dgst := range testLayers {
		fooCanonical, _ := reference.WithDigest(fooRepoName, dgst)
		option := WithMountFrom(fooCanonical)
		// ensure we can instrospect it
		createOpts := distribution.CreateOptions{}
		if err := option.Apply(&createOpts); err != nil {
			t.Fatalf("failed to apply MountFrom option: %v", err)
		}
		if !createOpts.Mount.ShouldMount || createOpts.Mount.From.String() != fooCanonical.String() {
			t.Fatalf("unexpected create options: %#+v", createOpts.Mount)
		}

		_, err := barRepo.Blobs(ctx).Create(ctx, WithMountFrom(fooCanonical))
		if err == nil {
			t.Fatalf("unexpected non-error while mounting from %q: %v", fooRepoName.String(), err)
		}
		if _, ok := err.(distribution.ErrBlobMounted); !ok {
			t.Fatalf("expected ErrMountFrom error, not %T: %v", err, err)
		}
	}
}
