package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	// Packages
	template "github.com/djthorpe/go-server/pkg/template"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path      string   `yaml:"path"`
	Templates string   `yaml:"templates"`
	Renderers []string `yaml:"renderers"`
	Default   string   `yaml:"default"`
}

type templates struct {
	Config
	*template.Cache
	filefs    fs.FS
	log       Logger
	mimetypes map[string]Renderer
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the template module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(templates)

	// Load configuration
	if err := provider.GetConfig(ctx, &this.Config); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Get logger
	if logger := provider.GetPlugin(ctx, "log").(Logger); logger == nil {
		provider.Print(ctx, "GetLogger: missing logger")
		return nil
	} else {
		this.log = logger
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

	//  Create cache for templates
	if stat, err := os.Stat(this.Config.Templates); err != nil || !stat.IsDir() {
		provider.Printf(ctx, "Invalid template path: %q", this.Config.Templates)
		return nil
	} else if cache, err := template.NewCache(os.DirFS(this.Config.Templates)); err != nil {
		provider.Printf(ctx, "NewCache: %q", err)
		return nil
	} else {
		this.Cache = cache
	}

	// Add handler for templates
	if err := provider.AddHandlerFunc(ctx, this.ServeHTTP); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Set renderers
	this.mimetypes = make(map[string]Renderer)
	for _, name := range this.Renderers {
		fmt.Println(name)
		if plugin := provider.GetPlugin(ctx, name); plugin == nil {
			provider.Printf(ctx, "Failed to load renderer: %q", name)
			return nil
		} else if renderer, ok := plugin.(Renderer); !ok {
			provider.Printf(ctx, "Failed to load renderer: %q", name)
			return nil
		} else {
			for _, mimetype := range renderer.Mimetypes() {
				if r, exists := this.mimetypes[mimetype]; exists {
					provider.Printf(ctx, "Duplicate renderer for %q, will be handled by %v", mimetype, r)
					return nil
				} else {
					this.mimetypes[mimetype] = renderer
				}
			}
		}
	}
	/*
		// Read templates
		if err := this.read(this.tmplfs); err != nil {
			provider.Print(ctx, "Failed to read templates: ", err)
			return nil
		}

		// Initiate the indexer
		if sqlite, ok := provider.GetModule(ctx, "sqlite").(nginx.SQPlugin); ok == false || sqlite == nil {
			provider.Print(ctx, "Failed to initialize indexer")
			return nil
		} else if err := this.indexer.Init(sqlite); err != nil {
			provider.Print(ctx, "Failed to initialize indexer: ", err)
			return nil
		}
	*/

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
	if len(this.Renderers) > 0 {
		str += fmt.Sprintf(" renderers=%q", this.Renderers)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "template"
}

func (this *templates) Run(ctx context.Context, _ Provider) error {
	<-ctx.Done()
	return nil
}
