package compiler

import (
	"math"
	"math/rand"

	"github.com/hilthontt/lotus/object"
)

func mathPackage() *object.Package {
	return &object.Package{
		Name: "Math",
		Functions: map[string]object.PackageFunction{

			"sqrt": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Sqrt(toFloat64(args[0]))}
			},

			"abs": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				switch v := args[0].(type) {
				case *object.Integer:
					val := v.Value
					if val < 0 {
						val = -val
					}
					return &object.Integer{Value: val}
				case *object.Float:
					return &object.Float{Value: math.Abs(v.Value)}
				}
				return &object.Nil{}
			},

			"floor": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(math.Floor(toFloat64(args[0])))}
			},

			"ceil": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(math.Ceil(toFloat64(args[0])))}
			},

			"round": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(math.Round(toFloat64(args[0])))}
			},

			"pow": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Pow(toFloat64(args[0]), toFloat64(args[1]))}
			},

			"max": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, b := toFloat64(args[0]), toFloat64(args[1])
				if a > b {
					return args[0]
				}
				return args[1]
			},

			"min": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, b := toFloat64(args[0]), toFloat64(args[1])
				if a < b {
					return args[0]
				}
				return args[1]
			},

			"pi": func(args ...object.Object) object.Object {
				return &object.Float{Value: math.Pi}
			},

			"e": func(args ...object.Object) object.Object {
				return &object.Float{Value: math.E}
			},

			// Math.log(x) -> float  (natural log)
			"log": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Log(toFloat64(args[0]))}
			},

			// Math.log2(x) -> float
			"log2": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Log2(toFloat64(args[0]))}
			},

			// Math.log10(x) -> float
			"log10": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Log10(toFloat64(args[0]))}
			},

			// Math.sin(x) -> float
			"sin": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Sin(toFloat64(args[0]))}
			},

			// Math.cos(x) -> float
			"cos": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Cos(toFloat64(args[0]))}
			},

			// Math.tan(x) -> float
			"tan": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Tan(toFloat64(args[0]))}
			},

			// Math.asin(x) -> float
			"asin": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Asin(toFloat64(args[0]))}
			},

			// Math.acos(x) -> float
			"acos": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Acos(toFloat64(args[0]))}
			},

			// Math.atan(x) -> float
			"atan": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Atan(toFloat64(args[0]))}
			},

			// Math.atan2(y, x) -> float
			"atan2": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Atan2(toFloat64(args[0]), toFloat64(args[1]))}
			},

			// Math.random() -> float  (0.0 to 1.0)
			"random": func(args ...object.Object) object.Object {
				return &object.Float{Value: rand.Float64()}
			},

			// Math.randomInt(min, max) -> int  (inclusive range)
			"randomInt": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				min, ok1 := args[0].(*object.Integer)
				max, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				if max.Value <= min.Value {
					return min
				}
				return &object.Integer{Value: min.Value + rand.Int63n(max.Value-min.Value+1)}
			},

			// Math.clamp(x, min, max) -> number
			"clamp": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				x := toFloat64(args[0])
				mn := toFloat64(args[1])
				mx := toFloat64(args[2])
				if x < mn {
					return args[1]
				}
				if x > mx {
					return args[2]
				}
				return args[0]
			},

			// Math.isNaN(x) -> bool
			"isNaN": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				f, ok := args[0].(*object.Float)
				if !ok {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: math.IsNaN(f.Value)}
			},

			// Math.isInf(x) -> bool
			"isInf": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				f, ok := args[0].(*object.Float)
				if !ok {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: math.IsInf(f.Value, 0)}
			},

			// Math.hypot(a, b) -> float  (sqrt(a²+b²))
			"hypot": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				return &object.Float{Value: math.Hypot(toFloat64(args[0]), toFloat64(args[1]))}
			},

			// Math.degrees(radians) -> float
			"degrees": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: toFloat64(args[0]) * 180.0 / math.Pi}
			},

			// Math.radians(degrees) -> float
			"radians": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Float{Value: toFloat64(args[0]) * math.Pi / 180.0}
			},

			// Math.gcd(a, b) -> int
			"gcd": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, ok1 := args[0].(*object.Integer)
				b, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				x, y := a.Value, b.Value
				if x < 0 {
					x = -x
				}
				if y < 0 {
					y = -y
				}
				for y != 0 {
					x, y = y, x%y
				}
				return &object.Integer{Value: x}
			},

			// Math.lcm(a, b) -> int
			"lcm": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, ok1 := args[0].(*object.Integer)
				b, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				x, y := a.Value, b.Value
				if x == 0 || y == 0 {
					return &object.Integer{Value: 0}
				}
				ax, ay := x, y
				if ax < 0 {
					ax = -ax
				}
				if ay < 0 {
					ay = -ay
				}
				gx, gy := ax, ay
				for gy != 0 {
					gx, gy = gy, gx%gy
				}
				return &object.Integer{Value: ax / gx * ay}
			},

			// Math.inf() -> float  (positive infinity)
			"inf": func(args ...object.Object) object.Object {
				return &object.Float{Value: math.Inf(1)}
			},

			// Math.nan() -> float
			"nan": func(args ...object.Object) object.Object {
				return &object.Float{Value: math.NaN()}
			},
		},
	}
}

func toFloat64(obj object.Object) float64 {
	switch v := obj.(type) {
	case *object.Integer:
		return float64(v.Value)
	case *object.Float:
		return v.Value
	}
	return 0
}
