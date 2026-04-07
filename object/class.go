package object

import (
	"bytes"
	"fmt"
	"strings"
)

// Class represents a Lotus class definition.
type Class struct {
	Name       string
	Methods    map[string]*Closure
	SuperClass *Class
}

func (c *Class) Type() ObjectType {
	return CLASS_OBJ
}

func (c *Class) Inspect() string {
	return fmt.Sprintf("<class %s>", c.Name)
}

// LookupMethod walks the class hierarchy to find a method by name.
func (c *Class) LookupMethod(name string) (*Closure, bool) {
	if m, ok := c.Methods[name]; ok {
		return m, true
	}
	if c.SuperClass != nil {
		return c.SuperClass.LookupMethod(name)
	}
	return nil, false
}

// Instance is a runtime instance of a Lotus class.
type Instance struct {
	Class  *Class
	Fields map[string]Object
}

func (i *Instance) Type() ObjectType {
	return INSTANCE_OBJ
}

func (i *Instance) Inspect() string {
	var out bytes.Buffer
	pairs := make([]string, 0, len(i.Fields))
	for k, v := range i.Fields {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.Inspect()))
	}
	out.WriteString(fmt.Sprintf("<%s {", i.Class.Name))
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}>")
	return out.String()
}

// SuperAccessor is the value of the 'super' expression inside a method.
// It holds the actual receiver (self) and the superclass to start method
// lookup from, enabling correct multi-level inheritance.
type SuperAccessor struct {
	Self       *Instance
	SuperClass *Class
}

func (sa *SuperAccessor) Type() ObjectType {
	return SUPER_OBJ
}

func (sa *SuperAccessor) Inspect() string {
	return "<super>"
}
