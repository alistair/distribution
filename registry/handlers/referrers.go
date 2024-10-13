package handlers

import (
	"encoding/json"
	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/internal/dcontext"
	"github.com/gorilla/handlers"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
	"net/http"
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

	manifestList := make([]distribution.Manifest, 0)

	//TODO Implement filtering my artifactType
	err = manifestEnumerator.EnumerateReferrer(rh.Context, rh.Digest, "", func(manifest distribution.Manifest) error {
		manifestList = append(manifestList, manifest)
		return nil
	})
	if err != nil {
		return
	}

	referrersApiResponse := responseIndex{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    v1.MediaTypeImageIndex,
		ArtifactType: "",
		Manifests:    manifestList,
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(referrersApiResponse); err != nil {
		return
	}
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
