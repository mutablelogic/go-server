package handler

import (
	"context"
	"errors"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	cert "github.com/mutablelogic/go-server/pkg/cert"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterCert(ctx context.Context, router server.HTTPRouter, prefix string, certmanager *cert.CertManager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "cert"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodPost:
			_ = certCreate(w, r, certmanager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "cert/{cert}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodDelete, http.MethodGet)

		// Parse request
		cert := r.PathValue("cert")
		if cert == "" {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing certificate name"))
			return
		}

		// Handle request
		switch r.Method {
		case http.MethodGet:
			_ = certGet(w, r, certmanager, cert)
		case http.MethodDelete:
			_ = certDelete(w, r, certmanager, cert)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func certCreate(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager) error {
	// Parse request
	var req schema.CertCreateMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the options for the certificate
	opts := []cert.Opt{}
	if req.CommonName != "" {
		opts = append(opts, cert.WithCommonName(req.CommonName))
	}
	if req.KeyType != "" {
		opts = append(opts, cert.WithKeyType(req.KeyType))
	} else {
		opts = append(opts, cert.WithDefaultKeyType())
	}
	if req.Expiry != 0 {
		opts = append(opts, cert.WithExpiry(req.Expiry))
	}
	if req.IsCA {
		opts = append(opts, cert.WithCA())
	}
	if req.Signer != "" {
		// Get signing certificate (needs to be a CA)
		signer, err := certmanager.GetCert(r.Context(), req.Signer)
		if errors.Is(err, httpresponse.ErrNotFound) || signer == nil {
			return httpresponse.Error(w, httpresponse.ErrNotFound.With("signer does not exist"), req.Signer)
		} else if err != nil {
			return httpresponse.Error(w, err)
		} else if !signer.IsCA {
			return httpresponse.Error(w, httpresponse.ErrBadRequest.With("signer is not a CA"), req.Signer)
		} else if signercert, err := cert.FromMeta(signer); err != nil {
			return httpresponse.Error(w, err)
		} else {
			opts = append(opts, cert.WithSigner(signercert))
		}
	}

	// Create a certificate with options
	cert, err := certmanager.CreateCert(r.Context(), req.Name, opts...)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), cert)
}

func certGet(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager, name string) error {
	response, err := certmanager.GetCert(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func certDelete(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager, name string) error {
	_, err := certmanager.DeleteCert(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}
