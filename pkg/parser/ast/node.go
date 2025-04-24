package ast

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Type int

type Node interface {
	// Return the type of node
	Type() Type

	// Parent node
	Parent() Node

	// Return children of the node
	Children() []Node

	// Return the underlying value, or nil if not applicable
	Value() any

	// Append a child node
	AppendChild(Node)
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Value types
	Null   Type = iota // Null value
	Ident              // Identifier
	String             // String literal
	Number             // Number literal
	Bool               // Boolean literal
	Array              // Array of values
	Dict               // Dict of values
)

func (t Type) String() string {
	switch t {
	case Null:
		return "Null"
	case Ident:
		return "Ident"
	case String:
		return "String"
	case Number:
		return "Number"
	case Bool:
		return "Bool"
	case Array:
		return "Array"
	case Dict:
		return "Dict"
	default:
		return "Unknown"
	}
}
