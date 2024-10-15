package storage

import (
	"context"
	p "path"

	storagedriver "github.com/distribution/distribution/v3/registry/storage/driver"

	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"
)

// ReferenceService is a service to manage internal links from subjects back to
// their referrers.
type ReferenceService interface {
	// Link creates a link from a subject back to a referrer
	Link(ctx context.Context, mediaType string, referrer, subject digest.Digest) error

	Enumerate(ctx context.Context, subject digest.Digest, mediaType string, ingestor func(digest.Digest) error) error
}

type referenceHandler struct {
	*blobStore
	repository distribution.Repository
	pathFn     func(name, mediaType string, reference, artifact_subject_must_be_manifest digest.Digest) (string, error)
	rootPathFn func(name, mediaType string, subject digest.Digest) (string, error)
}

func (r *referenceHandler) Link(ctx context.Context, artifactType string, referrer, subject digest.Digest) error {
	path, err := r.pathFn(r.repository.Named().Name(), artifactType, referrer, subject)
	if err != nil {
		return err
	}

	return r.blobStore.link(ctx, path, referrer)
}

func (r *referenceHandler) Enumerate(ctx context.Context, subject digest.Digest, mediaType string, ingestor func(digest.Digest) error) error {
	path, err := r.rootPathFn(r.repository.Named().Name(), mediaType, subject)
	if err != nil {
		return err
	}

	err = r.blobStore.driver.Walk(ctx, path, func(fileInfo storagedriver.FileInfo) error {
		// check if it's a link
		_, fileName := p.Split(fileInfo.Path())
		if fileName == "link" {
			d, err := r.blobStore.readlink(ctx, fileInfo.Path())
			if err != nil {
				return err
			}
			return ingestor(d)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// subjectReferrerLinkPath provides the path to the subject's referrer link
func subjectReferrerLinkPath(name, mediaType string, referrer, subject digest.Digest) (string, error) {
	return pathFor(subjectReferrerLinkPathSpec{name: name, mediaType: mediaType, referrer: referrer, subject: subject})
}

func subjectReferrerRootPath(name, mediaType string, subject digest.Digest) (string, error) {
	return pathFor(subjectReferrerRootPathSpec{name: name, mediaType: mediaType, subject: subject})
}
