package codegen

// VendorCodegen describes a vendor library's contributions to the code generation layer.
type VendorCodegen struct {
	// ModuleFuncs maps module name → (klang func name → C func name).
	ModuleFuncs map[string]map[string]string
	// ModuleConstants maps module name → (klang const name → C const name).
	ModuleConstants map[string]map[string]string
	// ConstNamespaces maps namespace name → (klang member → C constant).
	ConstNamespaces map[string]map[string]string
	// ValueTypes are C type names that are passed by value (not pointer).
	ValueTypes []string
	// TypeMap maps Klang type name → C type name (for resolveType).
	TypeMap map[string]string
	// ConstructorTypes are type names that use struct literal constructors (e.g. Color, Rectangle).
	ConstructorTypes []string
	// ReturnTypes maps function name → C return type (for inferring return types).
	ReturnTypes map[string]string
	// DetectIdents are identifier names that trigger vendor detection
	// (e.g. "rl", "Colors", "Key" trigger raylib detection).
	DetectIdents []string
	// DetectModule is the module name that triggers detection (e.g. "rl").
	DetectModule string
}

var vendorCodegens []*VendorCodegen

// RegisterVendorCodegen registers a vendor library with the codegen layer.
func RegisterVendorCodegen(v *VendorCodegen) {
	vendorCodegens = append(vendorCodegens, v)
}

// Precomputed lookup tables, populated by mergeVendors().
var (
	vendorValueTypes      map[string]bool
	vendorTypeMap         map[string]string
	vendorConstructorTypes map[string]bool
	vendorReturnTypes     map[string]string
	vendorDetectIdents    map[string]bool
	vendorDetectModules   map[string]bool
)

// vendorMerged tracks whether vendor data has been merged into the core maps.
var vendorMerged bool

// mergeVendors merges all vendor registrations into the core codegen maps.
// Called lazily on first use.
func mergeVendors() {
	if vendorMerged {
		return
	}
	vendorMerged = true

	vendorValueTypes = make(map[string]bool)
	vendorTypeMap = make(map[string]string)
	vendorConstructorTypes = make(map[string]bool)
	vendorReturnTypes = make(map[string]string)
	vendorDetectIdents = make(map[string]bool)
	vendorDetectModules = make(map[string]bool)

	for _, v := range vendorCodegens {
		for mod, funcs := range v.ModuleFuncs {
			if stdlibModuleFuncs[mod] == nil {
				stdlibModuleFuncs[mod] = make(map[string]string)
			}
			for k, cFunc := range funcs {
				stdlibModuleFuncs[mod][k] = cFunc
			}
			stdlibModules[mod] = true
		}
		for mod, consts := range v.ModuleConstants {
			if stdlibModuleConstants[mod] == nil {
				stdlibModuleConstants[mod] = make(map[string]string)
			}
			for k, cConst := range consts {
				stdlibModuleConstants[mod][k] = cConst
			}
		}
		for ns, members := range v.ConstNamespaces {
			if stdlibConstNamespaces[ns] == nil {
				stdlibConstNamespaces[ns] = make(map[string]string)
			}
			for k, cConst := range members {
				stdlibConstNamespaces[ns][k] = cConst
			}
		}
		for _, vt := range v.ValueTypes {
			vendorValueTypes[vt] = true
		}
		for k, cType := range v.TypeMap {
			vendorTypeMap[k] = cType
		}
		for _, ct := range v.ConstructorTypes {
			vendorConstructorTypes[ct] = true
		}
		for k, rt := range v.ReturnTypes {
			vendorReturnTypes[k] = rt
		}
		for _, ident := range v.DetectIdents {
			vendorDetectIdents[ident] = true
		}
		if v.DetectModule != "" {
			vendorDetectModules[v.DetectModule] = true
		}
	}
}
