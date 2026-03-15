package analysis

import (
	"strings"

	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

// SignatureResult holds signature help information.
type SignatureResult struct {
	Label           string
	Parameters      []ParamInfo
	ActiveParameter int
}

// SignatureHelp returns signature information for the function/method call at the given position.
func (d *Document) SignatureHelp(line, col int) *SignatureResult {
	if d.AST == nil {
		return nil
	}

	// Walk backward from cursor to find the opening '(' and the function name
	funcName, objName, activeParam := d.findCallContext(line, col)
	if funcName == "" {
		return nil
	}

	// Resolve the method/function signature
	if objName != "" {
		// Member call: obj.method(
		return d.signatureForMember(objName, funcName, activeParam, line)
	}

	// Bare call: could be a class method, stdlib with-module func, or constructor
	return d.signatureForBare(funcName, activeParam, line)
}

// findCallContext walks backward from the cursor to find the function being called
// and how many commas precede the cursor (activeParam).
func (d *Document) findCallContext(line, col int) (funcName, objName string, activeParam int) {
	// Collect tokens before cursor
	var prevTokens []lexer.Token
	for _, t := range d.Tokens {
		if t.Type == lexer.TOKEN_NEWLINE || t.Type == lexer.TOKEN_INDENT || t.Type == lexer.TOKEN_EOF {
			continue
		}
		if t.Line > line || (t.Line == line && t.Col >= col) {
			break
		}
		prevTokens = append(prevTokens, t)
	}

	if len(prevTokens) == 0 {
		return "", "", 0
	}

	// Walk backward, counting commas and matching parens to find the '('
	depth := 0
	commas := 0
	for i := len(prevTokens) - 1; i >= 0; i-- {
		t := prevTokens[i]
		switch t.Type {
		case lexer.TOKEN_RPAREN:
			depth++
		case lexer.TOKEN_LPAREN:
			if depth > 0 {
				depth--
			} else {
				// Found the matching '(' — the token before it is the function name
				if i > 0 && prevTokens[i-1].Type == lexer.TOKEN_IDENT {
					funcName = prevTokens[i-1].Value
					// Check for dot before the function name (member call)
					if i > 1 && (prevTokens[i-2].Type == lexer.TOKEN_DOT || prevTokens[i-2].Type == lexer.TOKEN_QUESTION_DOT) {
						if i > 2 && prevTokens[i-3].Type == lexer.TOKEN_IDENT {
							objName = prevTokens[i-3].Value
						}
					}
					return funcName, objName, commas
				}
				return "", "", 0
			}
		case lexer.TOKEN_COMMA:
			if depth == 0 {
				commas++
			}
		}
	}

	return "", "", 0
}

func (d *Document) signatureForMember(objName, methodName string, activeParam int, line int) *SignatureResult {
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	// Check stdlib modules
	if sigs, ok := StdlibModuleSignatures[objName]; ok {
		for _, sig := range sigs {
			if sig.Name == methodName {
				return buildSignatureFromDetail(sig.Detail, activeParam)
			}
		}
	}

	// Check event methods
	cls, _ := d.FindEnclosingClass(line)
	if cls != nil {
		if ev := d.findEventByName(cls, methodName); ev == nil {
			// Check if objName is an event
			if ev := d.findEventByName(cls, objName); ev != nil && methodName == "emit" {
				return buildEventEmitSignature(ev, activeParam)
			}
		}
	}

	// Resolve object type
	typeName := d.resolveIdentType(objName, line)
	if typeName == "" && objName == "this" {
		_, fullName := d.FindEnclosingClass(line)
		typeName = fullName
	}
	if typeName == "" {
		return nil
	}

	return d.findMethodSignature(typeName, methodName, activeParam, classes)
}

func (d *Document) findMethodSignature(typeName, methodName string, activeParam int, classes map[string]*parser.ClassDecl) *SignatureResult {
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}

	for _, m := range cls.Methods {
		if m.Name == methodName {
			return buildSignatureFromMethod(m, activeParam)
		}
	}

	// Check parent
	if cls.Parent != "" {
		return d.findMethodSignature(cls.Parent, methodName, activeParam, classes)
	}
	return nil
}

func (d *Document) signatureForBare(funcName string, activeParam int, line int) *SignatureResult {
	// Check with-module functions
	cls, _ := d.FindEnclosingClass(line)
	if cls != nil {
		method := d.FindEnclosingMethod(cls, line)
		if method != nil {
			mods := CollectWithModulesAtLine(method, line)
			for _, mod := range mods {
				if sigs, ok := StdlibModuleSignatures[mod]; ok {
					for _, sig := range sigs {
						if sig.Name == funcName {
							return buildSignatureFromDetail(sig.Detail, activeParam)
						}
					}
				}
			}
		}
	}

	// Check class methods (calling own method)
	if cls != nil {
		for _, m := range cls.Methods {
			if m.Name == funcName {
				return buildSignatureFromMethod(m, activeParam)
			}
		}
	}

	// Check constructors (e.g., Ball(), Color())
	classes := d.GetClasses()
	if classes != nil {
		for _, c := range classes {
			if c.Name == funcName && c.Constructor != nil {
				return buildSignatureFromConstructor(c, activeParam)
			}
		}
	}

	// Check built-in constructors
	builtins := map[string]string{
		"vec2":      "vec2(x:float, y:float)",
		"vec3":      "vec3(x:float, y:float, z:float)",
		"vec4":      "vec4(x:float, y:float, z:float, w:float)",
		"quat":      "quat(x:float, y:float, z:float, w:float)",
		"Color":     "Color(r:int, g:int, b:int, a:int)",
		"Rectangle": "Rectangle(x:float, y:float, width:float, height:float)",
	}
	if sig, ok := builtins[funcName]; ok {
		return buildSignatureFromDetail(sig, activeParam)
	}

	return nil
}

func buildSignatureFromMethod(m *parser.MethodDecl, activeParam int) *SignatureResult {
	var params []string
	var paramInfos []ParamInfo
	for _, p := range m.Params {
		ktype := typeExprToString(p.TypeExpr)
		s := p.Name
		if ktype != "" {
			s += ":" + ktype
		}
		params = append(params, s)
		paramInfos = append(paramInfos, ParamInfo{Name: p.Name, KType: ktype})
	}

	label := m.Name + "(" + strings.Join(params, ", ") + ")"
	if m.ReturnType != nil {
		label += ":" + typeExprToString(m.ReturnType)
	}

	return &SignatureResult{
		Label:           label,
		Parameters:      paramInfos,
		ActiveParameter: activeParam,
	}
}

func buildSignatureFromConstructor(cls *parser.ClassDecl, activeParam int) *SignatureResult {
	var params []string
	var paramInfos []ParamInfo
	for _, p := range cls.Constructor.Params {
		ktype := typeExprToString(p.TypeExpr)
		s := p.Name
		if ktype != "" {
			s += ":" + ktype
		}
		params = append(params, s)
		paramInfos = append(paramInfos, ParamInfo{Name: p.Name, KType: ktype})
	}

	label := cls.Name + "(" + strings.Join(params, ", ") + ")"

	return &SignatureResult{
		Label:           label,
		Parameters:      paramInfos,
		ActiveParameter: activeParam,
	}
}

func buildEventEmitSignature(ev *parser.EventDecl, activeParam int) *SignatureResult {
	var params []string
	var paramInfos []ParamInfo
	for _, p := range ev.Params {
		ktype := typeExprToString(p.TypeExpr)
		s := p.Name
		if ktype != "" {
			s += ":" + ktype
		}
		params = append(params, s)
		paramInfos = append(paramInfos, ParamInfo{Name: p.Name, KType: ktype})
	}
	label := ev.Name + ".emit(" + strings.Join(params, ", ") + ")"
	return &SignatureResult{
		Label:           label,
		Parameters:      paramInfos,
		ActiveParameter: activeParam,
	}
}

func buildSignatureFromDetail(detail string, activeParam int) *SignatureResult {
	// detail is like "func_name(param1:type1, param2:type2):return"
	// Parse out the parameters
	parenStart := strings.Index(detail, "(")
	parenEnd := strings.LastIndex(detail, ")")
	if parenStart < 0 || parenEnd < 0 {
		return &SignatureResult{Label: detail, ActiveParameter: activeParam}
	}

	paramStr := detail[parenStart+1 : parenEnd]
	var paramInfos []ParamInfo
	if paramStr != "" {
		parts := strings.Split(paramStr, ", ")
		for _, p := range parts {
			colonIdx := strings.Index(p, ":")
			if colonIdx >= 0 {
				paramInfos = append(paramInfos, ParamInfo{Name: p[:colonIdx], KType: p[colonIdx+1:]})
			} else {
				paramInfos = append(paramInfos, ParamInfo{Name: p})
			}
		}
	}

	return &SignatureResult{
		Label:           detail,
		Parameters:      paramInfos,
		ActiveParameter: activeParam,
	}
}
