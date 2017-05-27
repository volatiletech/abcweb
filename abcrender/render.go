package abcrender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/unrolled/render"
)

// Renderer implements template rendering methods.
// If you'd like to create a renderer that uses a different rendering engine
// opposed to standard text/templates or html/templates you can do so by
// implementing this interface.
type Renderer interface {
	Data(w io.Writer, status int, v []byte) error
	JSON(w io.Writer, status int, v interface{}) error
	Text(w io.Writer, status int, v string) error
	// HTML renders a HTML template. Example:
	// Assumes you have a template in ./templates called "home.tmpl"
	// $ mkdir -p templates && echo "<h1>Hello {{.}}</h1>" > templates/home.tmpl
	// HTML(w, http.StatusOK, "home", "World")
	HTML(w io.Writer, status int, name string, binding interface{}) error
	// HTMLWithLayout renders a HTML template using a different layout to the
	// one specified in your renderer's configuration. Example:
	// Example: HTMLWithLayout(w, http.StatusOK, "home", "World", "layout")
	HTMLWithLayout(w io.Writer, status int, name string, binding interface{}, layout string) error
}

// Render implements the HTML and HTMLWithLayout functions on the Renderer
// interface and imbeds the unrolled Render type to satisfy the rest of the interface.
// The custom HTML/HTMLWithLayout implementation is required due to the Render
// HTML function having a package-specific type for the layout string (Render.HTMLOptions).
// It's also required to wrap the AssetsManifest for the template function helpers.
type Render struct {
	*render.Render
	assetsManifest map[string]string
}

// HTML renders a HTML template by calling unrolled Render package's HTML function
func (r *Render) HTML(w io.Writer, status int, name string, binding interface{}) error {
	return r.Render.HTML(w, status, name, binding)
}

// HTMLWithLayout renders a HTML template using a specified layout file by calling
// unrolled Render package's HTML function with a HTMLOptions argument
func (r *Render) HTMLWithLayout(w io.Writer, status int, name string, binding interface{}, layout string) error {
	return r.Render.HTML(w, status, name, binding, render.HTMLOptions{Layout: layout})
}

// New returns a new Render with AssetsManifest and Render set
func New(opts render.Options, manifest map[string]string) Renderer {
	return &Render{
		Render:         render.New(opts),
		assetsManifest: manifest,
	}
}

// GetManifest reads the manifest.json file in the public assets folder
// and returns a map of its mappings. Returns error if manifest.json not found.
func GetManifest(publicPath string) (map[string]string, error) {
	manifestPath := filepath.Join(publicPath, "assets", "manifest.json")
	contents, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	if len(contents) == 0 {
		return nil, errors.New("manifest.json is empty")
	}

	manifest := map[string]string{}
	err = json.Unmarshal(contents, &manifest)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal manifest.json")
	}
	if len(manifest) == 0 {
		return nil, errors.New("manifest.json has no file mappings")
	}

	return manifest, nil
}

// AppHelpers takes in the assets manifest and returns a map of the template
// helper functions.
func AppHelpers(manifest map[string]string) template.FuncMap {
	return template.FuncMap{
		"liveReload": liveReloadHelper,
		"jsPath":     func(relpath string) string { return manifestHelper("js", relpath, manifest) },
		"cssPath":    func(relpath string) string { return manifestHelper("css", relpath, manifest) },
		"imgPath":    func(relpath string) string { return manifestHelper("img", relpath, manifest) },
		"videoPath":  func(relpath string) string { return manifestHelper("video", relpath, manifest) },
		"audioPath":  func(relpath string) string { return manifestHelper("audio", relpath, manifest) },
		"fontPath":   func(relpath string) string { return manifestHelper("font", relpath, manifest) },
		"assetPath": func(relpath string) string {
			v, ok := manifest[relpath]
			if !ok {
				return "/assets/" + relpath
			}
			return "/assets/" + v
		},

		// wrap full asset paths in include tags
		"cssTag": cssTag,
		"jsTag":  jsTag,

		"joinPath": func(pieces ...string) string { return strings.Join(pieces, "/") },

		// return all javascript include tags for all twitter bootstrap js plugins
		// for the default bootstrap install.
		"jsBootstrap": jsBootstrap,
	}
}

func manifestHelper(typ string, relpath string, manifest map[string]string) string {
	v, ok := manifest[filepath.Join(typ, relpath)]
	if !ok {
		return fmt.Sprintf("/assets/%s/%s", typ, relpath)
	}
	return "/assets/" + v
}

// liveReloadHelper is a helper to include the livereload javascript file
// and pass through a host queryparam so that it works properly.
func liveReloadHelper(relpath string, host string) string {
	return fmt.Sprintf("/assets/js/%s?host=%s", relpath, host)
}

// cssTag wraps the asset path in a css link include tag
func cssTag(relpath string) template.HTML {
	return template.HTML(fmt.Sprintf("<link href=\"%s\" rel=\"stylesheet\">", relpath))
}

// jsTag wraps the asset path in a javascript script include tag
func jsTag(relpath string) template.HTML {
	return template.HTML(fmt.Sprintf("<script src=\"%s\"></script>", relpath))
}

// jsBootstrap returns all javascript include tags for all twitter bootstrap
// js plugins for the default generated bootstrap install.
func jsBootstrap() template.HTML {
	files := []string{
		"/assets/js/bootstrap/transition.js",
		"/assets/js/bootstrap/util.js",
		"/assets/js/bootstrap/alert.js",
		"/assets/js/bootstrap/button.js",
		"/assets/js/bootstrap/carousel.js",
		"/assets/js/bootstrap/collapse.js",
		"/assets/js/bootstrap/dropdown.js",
		"/assets/js/bootstrap/modal.js",
		"/assets/js/bootstrap/scrollspy.js",
		"/assets/js/bootstrap/tab.js",
		"/assets/js/bootstrap/tooltip.js",
		"/assets/js/bootstrap/popover.js",
	}

	buf := bytes.Buffer{}
	for _, file := range files {
		buf.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", file))
	}
	return template.HTML(buf.String())
}
