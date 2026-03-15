package analysis

// VendorLib describes a vendor library's contributions to the analysis layer.
type VendorLib struct {
	// Modules maps module name → function signatures (e.g. "rl" → [...]).
	Modules map[string][]StdlibFunc
	// ModuleConstants maps module name → constant names (if any).
	ModuleConstants map[string][]StdlibFunc
	// Namespaces maps namespace name → members (e.g. "Colors" → [...]).
	Namespaces map[string][]StdlibFunc
	// Types are type names contributed by this vendor (e.g. "Color", "Texture2D").
	Types []string
	// TypeMembers maps type name → completion items for fields/methods.
	TypeMembers map[string][]CompletionItem
	// TypeFieldTypes maps type name → field name → field type (for chained resolution).
	TypeFieldTypes map[string]map[string]string
	// BuiltinConstructors maps type name → constructor signature string.
	BuiltinConstructors map[string]string
	// BuiltinConstructorParams maps type name → ordered param types.
	BuiltinConstructorParams map[string][]string
	// CTypeMap maps C type name → Klang display type (for cTypeToKlang).
	CTypeMap map[string]string
	// BuiltinIdents are additional identifiers that are always valid.
	BuiltinIdents []string
}

var vendorLibs []*VendorLib

// RegisterVendor registers a vendor library with the analysis layer.
func RegisterVendor(v *VendorLib) {
	vendorLibs = append(vendorLibs, v)
}

// VendorModuleSignatures returns all vendor-contributed module signatures.
func VendorModuleSignatures() map[string][]StdlibFunc {
	m := make(map[string][]StdlibFunc)
	for _, v := range vendorLibs {
		for k, funcs := range v.Modules {
			m[k] = append(m[k], funcs...)
		}
	}
	return m
}

// VendorModuleConstantNames returns all vendor-contributed module constants.
func VendorModuleConstantNames() map[string][]StdlibFunc {
	m := make(map[string][]StdlibFunc)
	for _, v := range vendorLibs {
		for k, consts := range v.ModuleConstants {
			m[k] = append(m[k], consts...)
		}
	}
	return m
}

// VendorNamespaces returns all vendor-contributed namespaces.
func VendorNamespaces() map[string][]StdlibFunc {
	m := make(map[string][]StdlibFunc)
	for _, v := range vendorLibs {
		for k, members := range v.Namespaces {
			m[k] = append(m[k], members...)
		}
	}
	return m
}

// VendorTypes returns all vendor-contributed type names.
func VendorTypes() []string {
	var types []string
	for _, v := range vendorLibs {
		types = append(types, v.Types...)
	}
	return types
}

// VendorTypeMembers returns all vendor-contributed type member completions.
func VendorTypeMembers() map[string][]CompletionItem {
	m := make(map[string][]CompletionItem)
	for _, v := range vendorLibs {
		for k, items := range v.TypeMembers {
			m[k] = append(m[k], items...)
		}
	}
	return m
}

// VendorTypeFieldTypes returns all vendor-contributed type field type mappings.
func VendorTypeFieldTypes() map[string]map[string]string {
	m := make(map[string]map[string]string)
	for _, v := range vendorLibs {
		for typeName, fields := range v.TypeFieldTypes {
			if m[typeName] == nil {
				m[typeName] = make(map[string]string)
			}
			for k, v := range fields {
				m[typeName][k] = v
			}
		}
	}
	return m
}

// VendorBuiltinConstructors returns all vendor-contributed constructor signatures.
func VendorBuiltinConstructors() map[string]string {
	m := make(map[string]string)
	for _, v := range vendorLibs {
		for k, sig := range v.BuiltinConstructors {
			m[k] = sig
		}
	}
	return m
}

// VendorBuiltinConstructorParams returns all vendor-contributed constructor param types.
func VendorBuiltinConstructorParams() map[string][]string {
	m := make(map[string][]string)
	for _, v := range vendorLibs {
		for k, params := range v.BuiltinConstructorParams {
			m[k] = params
		}
	}
	return m
}

// VendorCTypeMap returns all vendor-contributed C type → Klang type mappings.
func VendorCTypeMap() map[string]string {
	m := make(map[string]string)
	for _, v := range vendorLibs {
		for k, klType := range v.CTypeMap {
			m[k] = klType
		}
	}
	return m
}

// VendorBuiltinIdents returns all vendor-contributed builtin identifiers.
func VendorBuiltinIdents() []string {
	var idents []string
	for _, v := range vendorLibs {
		idents = append(idents, v.BuiltinIdents...)
	}
	return idents
}

// VendorModuleNames returns all vendor-contributed module names.
func VendorModuleNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, v := range vendorLibs {
		for k := range v.Modules {
			if !seen[k] {
				seen[k] = true
				names = append(names, k)
			}
		}
	}
	return names
}

// VendorNamespaceNames returns all vendor-contributed namespace names.
func VendorNamespaceNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, v := range vendorLibs {
		for k := range v.Namespaces {
			if !seen[k] {
				seen[k] = true
				names = append(names, k)
			}
		}
	}
	return names
}
