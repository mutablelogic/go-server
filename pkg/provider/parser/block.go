package parser

import "github.com/mutablelogic/go-server/pkg/provider/meta"

const (
	variable = "variable"
	comment  = "//"
)

type Resource struct {
	Meta *meta.Meta
}

type Variable struct {
	Name        string
	Description string
	Default     any
}
