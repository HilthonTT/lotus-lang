package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
)

const (
	ENUM_DEF_OBJ     ObjectType = "ENUM_DEF"
	ENUM_VARIANT_OBJ ObjectType = "ENUM_VARIANT"
)

// EnumVariantDef describes the fields of a variant (set at compile time).
type EnumVariantDef struct {
	Fields []string
}

// EnumDef is the runtime representation of an enum type.
// Accessing Color.Red returns an EnumVariant.
// Accessing Shape.Circle returns a Builtin constructor.
type EnumDef struct {
	Name     string
	Variants map[string]*EnumVariantDef
}

func (e *EnumDef) Type() ObjectType {
	return ENUM_DEF_OBJ
}

func (e *EnumDef) Inspect() string {
	return fmt.Sprintf("<enum %s>", e.Name)
}

// EnumVariant is a runtime enum value, e.g. Color.Red or Shape.Circle{radius: 5}.
type EnumVariant struct {
	EnumName    string
	VariantName string
	Data        map[string]Object // nil for simple (fieldless) variants
}

func (v *EnumVariant) Type() ObjectType {
	return ENUM_VARIANT_OBJ
}

func (v *EnumVariant) Inspect() string {
	if len(v.Data) == 0 {
		return v.EnumName + "." + v.VariantName
	}
	var out bytes.Buffer
	out.WriteString(v.EnumName + "." + v.VariantName + "{")
	pairs := make([]string, 0, len(v.Data))
	for k, val := range v.Data {
		pairs = append(pairs, k+": "+val.Inspect())
	}
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// HashKey allows enum variants to be used as map keys and compared.
func (v *EnumVariant) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(v.EnumName + "." + v.VariantName))
	return HashKey{Type: v.Type(), Value: h.Sum64()}
}

// GetField allows accessing fields on data-carrying variants: shape.radius
func (v *EnumVariant) GetField(name string) Object {
	if v.Data == nil {
		return &Nil{}
	}
	if val, ok := v.Data[name]; ok {
		return val
	}
	return &Nil{}
}
