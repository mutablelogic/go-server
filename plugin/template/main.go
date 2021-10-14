package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	// Packages
	template "github.com/mutablelogic/go-server/pkg/template"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path      string `yaml:"path"`
	Templates string `yaml:"templates"`
	Default   string `yaml:"default"`
}

type templates struct {
	Config
	Renderer
	*template.ContentTypeDetect
	*template.Cache
	filefs fs.FS
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the template module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(templates)
	this.ContentTypeDetect = template.NewContentTypeDetect()

	// Load configuration
	if err := provider.GetConfig(ctx, &this.Config); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Get renderer
	if renderer, ok := provider.GetPlugin(ctx, "renderer").(Renderer); !ok {
		provider.Print(ctx, "GetPlugin: missing renderer")
		return nil
	} else {
		this.Renderer = renderer
	}

	// Make paths absolute
	if path, err := filepath.Abs(this.Path); err != nil {
		provider.Print(ctx, "Invalid file system path:", err)
		return nil
	} else {
		this.Path = path
	}
	if path, err := filepath.Abs(this.Config.Templates); err != nil {
		provider.Print(ctx, "Invalid template path:", err)
		return nil
	} else {
		this.Config.Templates = path
	}

	// Set filesystem for documents
	if stat, err := os.Stat(this.Path); err != nil || !stat.IsDir() {
		provider.Printf(ctx, "Invalid file system path: %q", this.Path)
		return nil
	} else {
		this.filefs = os.DirFS(this.Path)
	}

	// Create cache for templates
	if stat, err := os.Stat(this.Config.Templates); err != nil || !stat.IsDir() {
		provider.Printf(ctx, "Invalid template path: %q", this.Config.Templates)
		return nil
	} else if cache, err := template.NewCache(os.DirFS(this.Config.Templates)); err != nil {
		provider.Printf(ctx, "NewCache: %q", err)
		return nil
	} else {
		this.Cache = cache
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *templates) String() string {
	str := "<template"
	if this.Config.Templates != "" {
		str += fmt.Sprintf(" templates=%q", this.Config.Templates)
	}
	if this.Config.Path != "" {
		str += fmt.Sprintf(" path=%q", this.Config.Path)
	}
	if this.Cache != nil {
		str += fmt.Sprint(" ", this.Cache)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Serves rendered documents through templates.\n")
	fmt.Fprintln(w, "  Configuration:")
	fmt.Fprintln(w, "    path: <path>")
	fmt.Fprintln(w, "      Required, folder that contains the files to be served")
	fmt.Fprintln(w, "    templates: <path>")
	fmt.Fprintln(w, "      Required, folder that contains the templates")
	fmt.Fprintln(w, "    default: <filename>")
	fmt.Fprintln(w, "      Required, the template which is used for rendering")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "template"
}

func (t *templates) Run(ctx context.Context, provider Provider) error {
	// Add handler for templates
	if err := provider.AddHandlerFunc(ctx, t.ServeHTTP); err != nil {
		return err
	}

	// Wait for end
	<-ctx.Done()

	// Return success
	return nil
}
