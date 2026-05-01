package compiler

import "github.com/hilthontt/lotus/object"

// packageEntry binds a package name to its constructor in a single place.
// Add new packages here — order matters and must never change, only append.
type packageEntry struct {
	name string
	pkg  *object.Package
}

var registeredPackages = []packageEntry{
	{"Console", consolePackage()},
	{"Math", mathPackage()},
	{"OS", osPackage()},
	{"Task", taskPackage()},
	{"Array", arrayPackage()},
	{"String", stringPackage()},
	{"Time", timePackage()},
	{"Json", jsonPackage()},
	{"File", filePackage()},
	{"Regex", regexPackage()},
	{"HttpClient", httpClientPackage()},
	{"Env", envPackage()},
}

// BuiltinPackageOrder and BuiltinPackages are derived from registeredPackages.
// Never assign to these directly — they exist only for consumers that need
// one form or the other.
var (
	BuiltinPackageOrder []string
	BuiltinPackages     map[string]*object.Package
)

func init() {
	BuiltinPackageOrder = make([]string, len(registeredPackages))
	BuiltinPackages = make(map[string]*object.Package, len(registeredPackages))

	seen := make(map[string]bool, len(registeredPackages))
	for i, entry := range registeredPackages {
		if entry.name == "" {
			panic("compiler: package entry at index " + itoa(i) + " has empty name")
		}
		if entry.pkg == nil {
			panic("compiler: package " + entry.name + " has nil *object.Package")
		}
		if seen[entry.name] {
			panic("compiler: duplicate package name: " + entry.name)
		}
		seen[entry.name] = true
		BuiltinPackageOrder[i] = entry.name
		BuiltinPackages[entry.name] = entry.pkg
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for ; i > 0; i /= 10 {
		s = string(rune('0'+i%10)) + s
	}
	return s
}
