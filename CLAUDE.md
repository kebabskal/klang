# Klang Development Guidelines

## LSP Requirements

When fixing bugs or adding features, always update the LSP accordingly:

- **Always update the LSP** with any fixes or added features — never leave the LSP out of sync.
- **Follow the whole chain** — if a feature touches parsing, type-checking, etc., make sure the LSP handles it end-to-end.
- **Signature help** — ensure functions/methods provide proper signature help.
- **Hovers** — all symbols should show type info on hover.
- **Generic types should be expanded** — when displaying generic types in the LSP (hover, completion, signatures), substitute concrete type parameters rather than showing raw generic placeholders.
- **Ensure the LSP restarts cleanly** — after changes, verify the LSP server can be restarted and picked up by VS Code without requiring a full editor restart. If protocol or initialization changes are made, test that a simple "Restart Language Server" command is sufficient.
- **Rebuild and reinstall the VS Code extension** — after LSP or extension changes, rebuild and reinstall the extension so VS Code picks up the latest version.

## Error Messaging

Klang compiles to C, but the user should never have to think about that:

- **Never surface C compiler errors** to the end user. Catch and translate errors at the Klang level before they reach the C compilation stage.
- **Be descriptive and clear** — error messages should tell the user what went wrong and where, in terms of their Klang source code.
- **Don't be overly verbose** — keep messages concise. One or two sentences plus a source location is usually enough.
- **Use Klang terminology** — refer to Klang types, constructs, and concepts, not C implementation details.
