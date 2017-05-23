package abcserver

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/volatiletech/abcmiddleware"
	"github.com/volatiletech/abcrender"
	"github.com/volatiletech/abcweb/abcconfig"
	"go.uber.org/zap"
)

type NotFound struct {
	Templates NotFoundTemplates
	// The manifest file mappings
	AssetsManifest map[string]string
}

type MethodNotAllowed struct {
	Templates MethodNotAllowedTemplates
}

type NotFoundTemplates struct {
	NotFound            string
	InternalServerError string
}

type MethodNotAllowedTemplates struct {
	MethodNotAllowed string
}

func NewMethodNotAllowedHandler() *MethodNotAllowed {
	return &MethodNotAllowed{
		Templates: MethodNotAllowedTemplates{
			MethodNotAllowed: "errors/405",
		},
	}
}

func NewNotFoundHandler(manifest map[string]string) *NotFound {
	return &NotFound{
		Templates: NotFoundTemplates{
			NotFound:            "errors/404",
			InternalServerError: "errors/500",
		},
		AssetsManifest: manifest,
	}
}

// NotFound handler is called if the requested route or asset cannot be found.
// Since we cannot use Chi's FileServer because it does directory listings
// we have to serve static assets (public folder) from the NotFound handler.
//
// The NotFound handler works for assets in both /public and /public/assets
//
// The NotFound handler checks if the path has "/assets", and
// if found will attempt to retrieve the asset name from the compiled
// assets manifest file if in production mode. In development mode it will
// ignore manifest and attempt to serve the asset directly.
//
// For paths that aren't "/assets/X" it will attempt to serve the asset
// directly, if it exists.
//
// Assets that cannot be found will return 404.
func (n *NotFound) Handler(cfg abcconfig.ServerConfig, render abcrender.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the Request ID scoped logger
		log := abcmiddleware.Log(r)

		reqPath := r.URL.Path
		// Ensure path is rooted at / to prevent path traversal
		if reqPath[0] != '/' {
			reqPath = "/" + reqPath
		}

		// Sanitize the path to prevent traversal exploits
		reqPath = path.Clean(reqPath)

		// the path to the asset file on disk
		var fpath string

		// Set path to asset in /assets, potentially contained in manifest
		if strings.HasPrefix(reqPath, "/assets/") {
			fname := strings.TrimPrefix(reqPath, "/assets/")

			ok := false
			// Look up the gzip version of the asset in the manifest
			// if the browser accepts gzip encoding
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				fpath, ok = n.AssetsManifest[fname+".gz"]
				if ok {
					w.Header().Set("Content-Encoding", "gzip")
				}
			}

			// If cannot find gzip version, attempt to serve regular version
			if !ok {
				fpath, ok = n.AssetsManifest[fname]
			}

			// If cannot find regular version in manifest, attempt to serve
			// using filename directly requested from browser
			if !ok {
				fpath = fname
			}
			fpath = filepath.Join(cfg.PublicPath, "assets", fpath)
		} else { // Set path to regular non-manifest asset
			// Split on ? to discard any query parameters
			fpath = filepath.Join(cfg.PublicPath, reqPath)
		}

		stat, err := os.Stat(fpath)
		// If file doesn't exist, or there's no error and the path is a dir, then 404
		if os.IsNotExist(err) || (err == nil && stat.IsDir()) {
			if err := render.HTML(w, http.StatusNotFound, n.Templates.NotFound, nil); err != nil {
				panic(err)
			}
			return
		} else if err != nil { // if not a file not exist error then http 500
			log.Fatal("failed to stat asset",
				zap.String("request_uri", r.RequestURI),
				zap.String("file_path", fpath),
				zap.Error(err),
			)
			if err := render.HTML(w, http.StatusInternalServerError, n.Templates.InternalServerError, nil); err != nil {
				panic(err)
			}
			return
		}

		fh, err := os.Open(fpath)
		if err != nil {
			log.Fatal("failed to open asset",
				zap.String("request_uri", r.RequestURI),
				zap.String("file_path", fpath),
				zap.Error(err),
			)
			if err := render.HTML(w, http.StatusInternalServerError, n.Templates.InternalServerError, nil); err != nil {
				panic(err)
			}
			return
		}

		// Serve the asset
		http.ServeContent(w, r, reqPath, stat.ModTime(), fh)

		fh.Close()
		return
	}
}

// MethodNotAllowed handler is called when someone attempts an operation
// against a route that does not support that operation, for example
// attempting a POST against a route that only supports a GET.
func (m *MethodNotAllowed) NewMethodNotAllowedHandler(render abcrender.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the Request ID scoped logger
		log := abcmiddleware.Log(r)

		log.Warn("method not allowed",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Bool("tls", r.TLS != nil),
			zap.String("protocol", r.Proto),
			zap.String("host", r.Host),
			zap.String("remote_addr", r.RemoteAddr),
		)

		if err := render.HTML(w, http.StatusMethodNotAllowed, m.Templates.MethodNotAllowed, nil); err != nil {
			panic(err)
		}
	}
}
