package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path      string                 `yaml:"path"`
	Templates string                 `yaml:"templates"`
	Renderers map[string]interface{} `yaml:"renderers"`
}

type template struct {
	Config

	filefs fs.FS
	tmplfs fs.FS
	log    Logger
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the template module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(template)

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
	if path, err := filepath.Abs(this.Templates); err != nil {
		provider.Print(ctx, "Invalid template path:", err)
		return nil
	} else {
		this.Templates = path
	}

	// Set filesystem for documents
	if stat, err := os.Stat(this.Path); err != nil || !stat.IsDir() {
		provider.Printf(ctx, "Invalid file system path: %q", this.Path)
		return nil
	} else {
		this.filefs = os.DirFS(this.Path)
	}

	//  Set filesystem for templates
	if stat, err := os.Stat(this.Templates); err != nil || !stat.IsDir() {
		provider.Printf(ctx, "Invalid template path: %q", this.Templates)
		return nil
	} else {
		this.tmplfs = os.DirFS(this.Templates)
	}

	// Add handler for templates
	if err := provider.AddHandler(ctx, this); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// TODO: Set template functions

	/*
		// Set renderers
		for name, mimetypes := range this.Renderers {
			if plugin := provider.GetModule(ctx, name); plugin == nil {
				provider.Printf(ctx, "Failed to load renderer: %q", name)
				return nil
			} else if renderer, ok := plugin.(nginx.Renderer); ok == false || renderer == nil {
				provider.Printf(ctx, "Failed to load renderer: %q", name)
				return nil
			} else if mimetypes, ok := mimetypes.([]interface{}); ok == false {
				provider.Printf(ctx, "Failed to set renderer mimetypes: %q", name)
				return nil
			} else {
				for _, mimetype := range mimetypes {
					if err := this.renderers.Add(mimetype.(string), renderer); err != nil {
						provider.Printf(ctx, "Failed to set renderer for %q: %v", name, err)
						return nil
					}
				}
			}
		}

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

func (this *template) String() string {
	str := "<template"
	if this.Config.Templates != "" {
		str += fmt.Sprintf(" templates=%q", this.Config.Templates)
	}
	if this.Config.Path != "" {
		str += fmt.Sprintf(" path=%q", this.Config.Path)
	}
	for renderer, mimetypes := range this.Config.Renderers {
		str += fmt.Sprintf(" %q=%q", renderer, mimetypes)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "template"
}

func (this *template) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
