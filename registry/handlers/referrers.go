package handlers

import (
	"net/http"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/internal/dcontext"
	"github.com/distribution/distribution/v3/manifest/ocischema"
	"github.com/gorilla/handlers"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
)

// referrersDispatcher takes the request context and builds the
// appropriate handler for handling manifest requests.
func referrersDispatcher(ctx *Context, r *http.Request) http.Handler {
	referrersHandler := &referrersHandler{
		Context: ctx,
	}
	dgst, err := getDigest(ctx)
	if err != nil {
		log.Errorf("Invalid referrers digest: %s", dgst)
	} else {
		referrersHandler.Digest = dgst
	}

	mhandler := handlers.MethodHandler{
		http.MethodGet: http.HandlerFunc(referrersHandler.GetManifest),
	}

	return mhandler
}

// referrersHandle handles http operations on referrer requests.
type referrersHandler struct {
	*Context

	Digest digest.Digest
}

func (rh *referrersHandler) GetManifest(w http.ResponseWriter, r *http.Request) {
	dcontext.GetLogger(rh).Debug("GetReferrers")
	manifests, err := rh.Repository.Manifests(rh)
	manifestEnumerator := manifests.(distribution.ManifestReferrerEnumerator)
	if nil == manifestEnumerator {
		return
	}
	if err != nil {
		rh.Errors = append(rh.Errors, err)
		return
	}

	manifestList := make([]v1.Descriptor, 0)

	//TODO Implement filtering my artifactType
	err = manifestEnumerator.EnumerateReferrer(rh.Context, rh.Digest, "", func(desc v1.Descriptor) error {
		manifestList = append(manifestList, desc)
		return nil
	})
	if err != nil {
		return
	}

	s, err = ocischema.FromDescriptors(manifestList, nil)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", v1.MediaTypeImageIndex)
	w.Write(s.Payload())

	return
}

// Index references manifests for various platforms.
// This structure provides `application/vnd.oci.image.index.v1+json` mediatype when marshalled to JSON.
type responseIndex struct {
	specs.Versioned

	// MediaType specifies the type of this document data structure e.g. `application/vnd.oci.image.index.v1+json`
	MediaType string `json:"mediaType,omitempty"`

	// ArtifactType specifies the IANA media type of artifact when the manifest is used for an artifact.
	ArtifactType string `json:"artifactType,omitempty"`

	// Manifests references platform specific manifests.
	Manifests []distribution.Manifest `json:"manifests"`
}
