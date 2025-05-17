package parser

import "github.com/mutablelogic/go-server/pkg/provider/meta"

const (
	kwVariable    = "variable"
	kwDescription = "description"
	kwDefault     = "default"
	kwComment     = "//"
)

type Resource struct {
	Meta  *meta.Meta
	Label string
}

type Variable struct {
	Name        string
	Description string
	Default     any
}
