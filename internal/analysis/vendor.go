package analysis

import "sync"

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

// mergeVendorsOnce ensures vendor data is merged into core tables exactly once.
var mergeVendorsOnce sync.Once

// ensureVendorsMerged merges all registered vendor data into the core stdlib
// maps. Called lazily on first use, after all init() registrations have completed.
func ensureVendorsMerged() {
	mergeVendorsOnce.Do(func() {
		for _, v := range vendorLibs {
			for k, funcs := range v.Modules {
				StdlibModuleSignatures[k] = append(StdlibModuleSignatures[k], funcs...)
			}
			for k, consts := range v.ModuleConstants {
				StdlibModuleConstantNames[k] = append(StdlibModuleConstantNames[k], consts...)
			}
			for k, members := range v.Namespaces {
				StdlibNamespaces[k] = append(StdlibNamespaces[k], members...)
			}
			BuiltinTypes = append(BuiltinTypes, v.Types...)
			for k, items := range v.TypeMembers {
				BuiltinTypeMembers[k] = append(BuiltinTypeMembers[k], items...)
			}
			for typeName, fields := range v.TypeFieldTypes {
				if BuiltinTypeFieldTypes[typeName] == nil {
					BuiltinTypeFieldTypes[typeName] = make(map[string]string)
				}
				for fk, fv := range fields {
					BuiltinTypeFieldTypes[typeName][fk] = fv
				}
			}
			ModuleNames = append(ModuleNames, VendorModuleNames()...)
			NamespaceNames = append(NamespaceNames, VendorNamespaceNames()...)
			builtinIdents = append(builtinIdents, v.BuiltinIdents...)
			for k, klType := range v.CTypeMap {
				vendorCTypes[k] = klType
			}
		}
	})
}

// VendorBuiltinConstructors returns all vendor-contributed constructor signatures.
func VendorBuiltinConstructors() map[string]string {
	ensureVendorsMerged()
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
	ensureVendorsMerged()
	m := make(map[string][]string)
	for _, v := range vendorLibs {
		for k, params := range v.BuiltinConstructorParams {
			m[k] = params
		}
	}
	return m
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
