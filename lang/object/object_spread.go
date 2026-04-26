package object

const SPREAD_OBJ ObjectType = "SPREAD"

type SpreadValue struct {
	Elements []Object
}

func (s *SpreadValue) Type() ObjectType {
	return SPREAD_OBJ
}

func (s *SpreadValue) Inspect() string {
	return "[...spread]"
}
