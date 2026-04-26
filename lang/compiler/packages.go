package compiler

import "github.com/hilthontt/lotus/object"

// BuiltinPackageOrder defines the registration order so global indices
// are identical across every compiler.New() call.
var BuiltinPackageOrder = []string{
	"Console",
	"Math",
	"OS",
	"Task",
	"Array",
	"String",
	"Time",
	"Json",
	"File",
	"Regex",
}

// Add new packages here — they are automatically injected as globals.
var BuiltinPackages = map[string]*object.Package{
	"Console": consolePackage(),
	"Math":    mathPackage(),
	"OS":      osPackage(),
	"Task":    taskPackage(),
	"Array":   arrayPackage(),
	"String":  stringPackage(),
	"Time":    timePackage(),
	"Json":    jsonPackage(),
	"File":    filePackage(),
	"Regex":   regexPackage(),
}
