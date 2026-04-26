package compiler

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilthontt/lotus/object"
)

func filePackage() *object.Package {
	return &object.Package{
		Name: "File",
		Functions: map[string]object.PackageFunction{
			// File.exists(path) -> bool
			"exists": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				_, err := os.Stat(p.Value)
				return &object.Boolean{Value: err == nil}
			},
			// File.read(path) -> string | nil
			"read": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				data, err := os.ReadFile(p.Value)
				if err != nil {
					return &object.Nil{}
				}
				return &object.String{Value: string(data)}
			},
			// File.write(path, content) -> bool
			"write": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				p, ok1 := args[0].(*object.String)
				content, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Boolean{Value: false}
				}
				err := os.WriteFile(p.Value, []byte(content.Value), 0644)
				return &object.Boolean{Value: err == nil}
			},
			// File.append(path, content) -> bool
			"append": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				p, ok1 := args[0].(*object.String)
				content, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Boolean{Value: false}
				}
				f, err := os.OpenFile(p.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				defer f.Close()
				_, err = f.WriteString(content.Value)
				return &object.Boolean{Value: err == nil}
			},
			// File.delete(path) -> bool
			"delete": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				err := os.Remove(p.Value)
				return &object.Boolean{Value: err == nil}
			},
			// File.rename(old, new) -> bool
			"rename": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				old, ok1 := args[0].(*object.String)
				new_, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Boolean{Value: false}
				}
				err := os.Rename(old.Value, new_.Value)
				return &object.Boolean{Value: err == nil}
			},
			// File.copy(src, dst) -> bool
			"copy": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				src, ok1 := args[0].(*object.String)
				dst, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Boolean{Value: false}
				}
				in, err := os.Open(src.Value)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				defer in.Close()
				out, err := os.Create(dst.Value)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				defer out.Close()
				_, err = io.Copy(out, in)
				return &object.Boolean{Value: err == nil}
			},
			// File.mkdir(path) -> bool  (creates all parent directories)
			"mkdir": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				err := os.MkdirAll(p.Value, 0755)
				return &object.Boolean{Value: err == nil}
			},
			// File.listDir(path) -> array of strings
			"listDir": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				entries, err := os.ReadDir(p.Value)
				if err != nil {
					return &object.Nil{}
				}
				elems := make([]object.Object, len(entries))
				for i, e := range entries {
					elems[i] = &object.String{Value: e.Name()}
				}
				return &object.Array{Elements: elems}
			},

			// File.isDir(path) -> bool
			"isDir": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				info, err := os.Stat(p.Value)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: info.IsDir()}
			},

			// File.isFile(path) -> bool
			"isFile": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				info, err := os.Stat(p.Value)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: !info.IsDir()}
			},

			// File.size(path) -> int  (bytes, -1 on error)
			"size": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Integer{Value: -1}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Integer{Value: -1}
				}
				info, err := os.Stat(p.Value)
				if err != nil {
					return &object.Integer{Value: -1}
				}
				return &object.Integer{Value: info.Size()}
			},

			// File.extension(path) -> string  (e.g. ".go", "" if none)
			"extension": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.String{Value: ""}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.String{Value: ""}
				}
				return &object.String{Value: filepath.Ext(p.Value)}
			},

			// File.basename(path) -> string  (filename without directory)
			"basename": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.String{Value: ""}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.String{Value: ""}
				}
				return &object.String{Value: filepath.Base(p.Value)}
			},

			// File.dirname(path) -> string
			"dirname": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.String{Value: ""}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.String{Value: ""}
				}
				return &object.String{Value: filepath.Dir(p.Value)}
			},

			// File.join(...parts) -> string
			"join": func(args ...object.Object) object.Object {
				parts := make([]string, len(args))
				for i, a := range args {
					s, ok := a.(*object.String)
					if !ok {
						return &object.Nil{}
					}
					parts[i] = s.Value
				}
				return &object.String{Value: filepath.Join(parts...)}
			},

			// File.abs(path) -> string
			"abs": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				abs, err := filepath.Abs(p.Value)
				if err != nil {
					return &object.Nil{}
				}
				return &object.String{Value: abs}
			},

			// File.stem(path) -> string  (basename without extension)
			"stem": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.String{Value: ""}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.String{Value: ""}
				}
				base := filepath.Base(p.Value)
				ext := filepath.Ext(base)
				return &object.String{Value: strings.TrimSuffix(base, ext)}
			},

			// File.cwd() -> string  (current working directory)
			"cwd": func(args ...object.Object) object.Object {
				dir, err := os.Getwd()
				if err != nil {
					return &object.Nil{}
				}
				return &object.String{Value: dir}
			},

			// File.glob(pattern) -> array  (e.g. File.glob("*.lotus"))
			"glob": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				p, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				matches, err := filepath.Glob(p.Value)
				if err != nil {
					return &object.Nil{}
				}
				elems := make([]object.Object, len(matches))
				for i, m := range matches {
					elems[i] = &object.String{Value: m}
				}
				return &object.Array{Elements: elems}
			},
		},
	}
}
