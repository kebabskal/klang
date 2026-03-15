package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/klang-lang/klang/internal/analysis"
	"github.com/klang-lang/klang/internal/errs"
)

// --- JSON-RPC types ---

type jsonrpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- LSP types (minimal subset) ---

type InitializeParams struct {
	RootURI string `json:"rootUri"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync      int                    `json:"textDocumentSync"` // 1 = Full
	CompletionProvider    *CompletionOptions     `json:"completionProvider,omitempty"`
	HoverProvider         bool                   `json:"hoverProvider,omitempty"`
	DefinitionProvider    bool                   `json:"definitionProvider,omitempty"`
	SignatureHelpProvider *SignatureHelpOptions   `json:"signatureHelpProvider,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidChangeParams struct {
	TextDocument   VersionedTextDocumentID `json:"textDocument"`
	ContentChanges []ContentChange         `json:"contentChanges"`
}

type VersionedTextDocumentID struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type ContentChange struct {
	Text string `json:"text"`
}

type DidCloseParams struct {
	TextDocument TextDocumentID `json:"textDocument"`
}

type TextDocumentID struct {
	URI string `json:"uri"`
}

type CompletionParams struct {
	TextDocument TextDocumentID `json:"textDocument"`
	Position     Position       `json:"position"`
}

type HoverParams struct {
	TextDocument TextDocumentID `json:"textDocument"`
	Position     Position       `json:"position"`
}

type DefinitionParams struct {
	TextDocument TextDocumentID `json:"textDocument"`
	Position     Position       `json:"position"`
}

type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range   `json:"range"`
	Severity int     `json:"severity"` // 1=Error, 2=Warning
	Source   string  `json:"source"`
	Message  string  `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind,omitempty"`
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation *MarkupContent         `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label         [2]int         `json:"label"` // [start, end] offsets in signature label
	Documentation *MarkupContent `json:"documentation,omitempty"`
}

type SignatureHelpParams struct {
	TextDocument TextDocumentID `json:"textDocument"`
	Position     Position       `json:"position"`
}

// --- Server state ---

var (
	docs     = map[string]*analysis.Document{}
	rootPath string // workspace root from initialize
	writer   *bufio.Writer
	logger   *log.Logger
)

func main() {
	// Log to stderr for debugging
	logger = log.New(os.Stderr, "[kl-lsp] ", log.LstdFlags)
	writer = bufio.NewWriter(os.Stdout)

	reader := bufio.NewReader(os.Stdin)

	for {
		msg, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			logger.Printf("read error: %v", err)
			return
		}
		handleMessage(msg)
	}
}

func readMessage(r *bufio.Reader) (*jsonrpcMessage, error) {
	// Read headers
	contentLength := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no Content-Length header")
	}

	// Read body
	body := make([]byte, contentLength)
	_, err := io.ReadFull(r, body)
	if err != nil {
		return nil, err
	}

	var msg jsonrpcMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func sendResponse(id json.RawMessage, result interface{}) {
	// JSON-RPC requires "result" field in every response (even if null).
	// We marshal manually because omitempty would drop nil results.
	resultJSON, _ := json.Marshal(result)
	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":%s,"result":%s}`, string(id), string(resultJSON))
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	writer.WriteString(header)
	writer.WriteString(body)
	writer.Flush()
}

func sendNotification(method string, params interface{}) {
	msg := jsonrpcMessage{
		JSONRPC: "2.0",
		Method:  method,
	}
	if params != nil {
		raw, _ := json.Marshal(params)
		msg.Params = raw
	}
	sendMessage(msg)
}

func sendMessage(msg jsonrpcMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		logger.Printf("marshal error: %v", err)
		return
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	writer.WriteString(header)
	writer.Write(body)
	writer.Flush()
}

func handleMessage(msg *jsonrpcMessage) {
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("panic handling %s: %v", msg.Method, r)
			// If it was a request (has ID), send an error response
			if msg.ID != nil {
				sendResponse(msg.ID, nil)
			}
		}
	}()
	switch msg.Method {
	case "initialize":
		handleInitialize(msg)
	case "initialized":
		// no-op
	case "shutdown":
		sendResponse(msg.ID, nil)
	case "exit":
		os.Exit(0)
	case "textDocument/didOpen":
		handleDidOpen(msg)
	case "textDocument/didChange":
		handleDidChange(msg)
	case "textDocument/didClose":
		handleDidClose(msg)
	case "textDocument/completion":
		handleCompletion(msg)
	case "textDocument/hover":
		handleHover(msg)
	case "textDocument/definition":
		handleDefinition(msg)
	case "textDocument/signatureHelp":
		handleSignatureHelp(msg)
	default:
		// Unknown method — return method not found for requests (that have an ID)
		if msg.ID != nil {
			resp := jsonrpcMessage{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Error:   &jsonrpcError{Code: -32601, Message: "method not found: " + msg.Method},
			}
			sendMessage(resp)
		}
	}
}

func handleInitialize(msg *jsonrpcMessage) {
	var params InitializeParams
	json.Unmarshal(msg.Params, &params)
	if params.RootURI != "" {
		rootPath = uriToPath(params.RootURI)
		logger.Printf("initialize: rootPath=%s", rootPath)
	}

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", ":"},
			},
			HoverProvider:      true,
			DefinitionProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
		},
	}
	sendResponse(msg.ID, result)
}

func handleDidOpen(msg *jsonrpcMessage) {
	var params DidOpenParams
	json.Unmarshal(msg.Params, &params)

	uri := params.TextDocument.URI
	logger.Printf("didOpen: uri=%s path=%s len=%d", uri, uriToPath(uri), len(params.TextDocument.Text))
	doc := analyzeWithSiblings(uri, []byte(params.TextDocument.Text))
	logger.Printf("didOpen: tokens=%d classes=%d diags=%d ast=%v", len(doc.Tokens), len(doc.GetClasses()), len(doc.Diags), doc.AST != nil)
	docs[uri] = doc
	publishDiagnostics(uri, doc)
}

func handleDidChange(msg *jsonrpcMessage) {
	var params DidChangeParams
	json.Unmarshal(msg.Params, &params)

	uri := params.TextDocument.URI
	if len(params.ContentChanges) == 0 {
		return
	}
	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	doc := analyzeWithSiblings(uri, []byte(text))
	docs[uri] = doc
	publishDiagnostics(uri, doc)
}

func handleDidClose(msg *jsonrpcMessage) {
	var params DidCloseParams
	json.Unmarshal(msg.Params, &params)
	delete(docs, params.TextDocument.URI)

	// Clear diagnostics
	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []Diagnostic{},
	})
}

func handleCompletion(msg *jsonrpcMessage) {
	var params CompletionParams
	json.Unmarshal(msg.Params, &params)

	doc := docs[params.TextDocument.URI]
	if doc == nil {
		sendResponse(msg.ID, CompletionList{})
		return
	}

	// LSP positions are 0-based, our analysis uses 1-based
	items := doc.Complete(params.Position.Line+1, params.Position.Character+1)

	lspItems := make([]CompletionItem, len(items))
	for i, item := range items {
		lspItems[i] = CompletionItem{
			Label:      item.Label,
			Kind:       int(item.Kind),
			Detail:     item.Detail,
			InsertText: item.InsertText,
		}
	}

	sendResponse(msg.ID, CompletionList{Items: lspItems})
}

func handleHover(msg *jsonrpcMessage) {
	var params HoverParams
	json.Unmarshal(msg.Params, &params)

	doc := docs[params.TextDocument.URI]
	if doc == nil {
		sendResponse(msg.ID, nil)
		return
	}

	line := params.Position.Line + 1
	col := params.Position.Character + 1

	result := doc.Hover(line, col)
	if result == nil {
		sendResponse(msg.ID, nil)
		return
	}

	hover := Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: result.Content,
		},
	}
	if result.Line > 0 {
		hover.Range = &Range{
			Start: Position{Line: result.Line - 1, Character: result.Col - 1},
			End:   Position{Line: result.Line - 1, Character: result.EndCol - 1},
		}
	}
	sendResponse(msg.ID, hover)
}

func handleDefinition(msg *jsonrpcMessage) {
	var params DefinitionParams
	json.Unmarshal(msg.Params, &params)

	doc := docs[params.TextDocument.URI]
	if doc == nil {
		sendResponse(msg.ID, nil)
		return
	}

	result := doc.Definition(params.Position.Line+1, params.Position.Character+1)
	if result == nil {
		sendResponse(msg.ID, nil)
		return
	}

	// Use the URI from the definition result (may be a different file)
	defURI := params.TextDocument.URI
	if result.URI != "" {
		defURI = result.URI
	}

	loc := Location{
		URI: defURI,
		Range: Range{
			Start: Position{Line: result.Line - 1, Character: result.Col - 1},
			End:   Position{Line: result.Line - 1, Character: result.EndCol - 1},
		},
	}
	sendResponse(msg.ID, loc)
}

func handleSignatureHelp(msg *jsonrpcMessage) {
	var params SignatureHelpParams
	json.Unmarshal(msg.Params, &params)

	doc := docs[params.TextDocument.URI]
	if doc == nil {
		sendResponse(msg.ID, nil)
		return
	}

	result := doc.SignatureHelp(params.Position.Line+1, params.Position.Character+1)
	if result == nil {
		sendResponse(msg.ID, nil)
		return
	}

	// Build parameter label offsets from the signature label
	var lspParams []ParameterInformation
	label := result.Label
	for _, p := range result.Parameters {
		// Find the parameter substring in the label
		paramLabel := p.Name
		if p.KType != "" {
			paramLabel += ":" + p.KType
		}
		start := strings.Index(label, paramLabel)
		if start >= 0 {
			lspParams = append(lspParams, ParameterInformation{
				Label: [2]int{start, start + len(paramLabel)},
			})
		}
	}

	sig := SignatureInformation{
		Label:      label,
		Parameters: lspParams,
	}

	sendResponse(msg.ID, SignatureHelp{
		Signatures:      []SignatureInformation{sig},
		ActiveSignature: 0,
		ActiveParameter: result.ActiveParameter,
	})
}

// analyzeWithSiblings analyzes a file and enriches it with classes from sibling .k files.
// This enables cross-file type resolution, completion, hover, and go-to-definition.
func analyzeWithSiblings(uri string, text []byte) *analysis.Document {
	filePath := uriToPath(uri)
	doc := analysis.Analyze(filePath, text)
	if doc == nil || doc.Gen == nil {
		return doc
	}

	// Find sibling .k files in the same project directory
	siblings := findProjectKFiles(filePath)
	for _, sibPath := range siblings {
		// Skip the current file
		if filepath.Clean(sibPath) == filepath.Clean(filePath) {
			continue
		}

		// Check if this sibling is already open in the editor (use in-memory text)
		sibURI := pathToURI(sibPath)
		if openDoc, ok := docs[sibURI]; ok && openDoc.AST != nil {
			doc.Gen.AddFile(openDoc.AST)
			doc.AddSiblingFile(sibURI, openDoc.AST)
			continue
		}

		// Parse from disk
		src, err := os.ReadFile(sibPath)
		if err != nil {
			continue
		}
		sibDoc := analysis.Analyze(sibPath, src)
		if sibDoc != nil && sibDoc.AST != nil {
			doc.Gen.AddFile(sibDoc.AST)
			doc.AddSiblingFile(sibURI, sibDoc.AST)
		}
	}

	// Re-run semantic checks now that sibling classes are registered
	if doc.AST != nil {
		doc.Diags = nil
		doc.Diags = append(doc.Diags, doc.ParseDiags...)
		if len(doc.ParseDiags) == 0 {
			doc.Check()
		}
	}

	return doc
}

// findProjectKFiles finds all .k files in the same directory tree as the given file.
// Walks the file's directory and its subdirectories (not the entire workspace).
func findProjectKFiles(filePath string) []string {
	dir := filepath.Dir(filePath)

	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip build directories
		if info.IsDir() && (info.Name() == "build" || info.Name() == ".git" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".k") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func pathToURI(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	absPath = filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" {
		return "file:///" + absPath
	}
	return "file://" + absPath
}

func publishDiagnostics(uri string, doc *analysis.Document) {
	var lspDiags []Diagnostic
	for _, d := range doc.Diags {
		severity := 1 // Error
		if d.Kind == errs.Warning {
			severity = 2
		}
		lspDiag := Diagnostic{
			Range: Range{
				Start: Position{Line: max(0, d.Line-1), Character: max(0, d.Col-1)},
				End:   Position{Line: max(0, d.Line-1), Character: max(0, d.EndCol-1)},
			},
			Severity: severity,
			Source:   "klang",
			Message:  d.Message,
		}
		lspDiags = append(lspDiags, lspDiag)
	}
	if lspDiags == nil {
		lspDiags = []Diagnostic{} // empty array, not null
	}

	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	p := u.Path
	// On Windows, file URIs have paths like /C:/Users/...
	if runtime.GOOS == "windows" && len(p) > 2 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}
