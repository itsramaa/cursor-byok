# Design: Cross-OS Line Ending Normalization

## Architecture

### New File: `internal/agent/lineendings.go`

Single shared helper used by all write paths:

```go
func normalizeLineEndingsForFile(path, content string) string
```

**Detection algorithm:**
1. Collapse all `\r\n` → `\n` (clean baseline)
2. Read existing file bytes (if exists)
3. Sniff: first `\r\n` found → CRLF; first bare `\n` found → LF; no newlines → OS default
4. Apply: if CRLF detected, `\n` → `\r\n`; otherwise keep as `\n`

**OS fallback** (new files):
- `runtime.GOOS == "windows"` → CRLF
- Everything else → LF

### Call Sites Modified

| File | Function | Change |
|---|---|---|
| `tool_exec.go` | `Write` case (line 721-746) | Add `body = normalizeLineEndingsForFile(a.Path, body)` before `StreamContent` |
| `tool_exec.go` | `StrReplace` case (line 599-654) | Replace entire normalization block with `after = normalizeLineEndingsForFile(a.Path, after)` |
| `tool_exec.go` | `writeExecRequest` EditArgs (line 882-889) | Add `args.FileText = normalizeLineEndingsForFile(args.Path, args.FileText)` |
| `bugbot.go` | `writeBugBotExecRequest` EditArgs (line 282-290) | Same as above |
| `stream.js` (9router) | passthrough split | Add `\r` strip after `split("\n")` |

### Test Plan

Unit test `lineendings_test.go`:
- `TestDetectCRLF_WindowsFile` — existing `\r\n` file stays `\r\n`
- `TestDetectCRLF_LinuxFile` — existing `\n` file stays `\n`  
- `TestDetectCRLF_NewFileOnWindows` — new file gets `\r\n`
- `TestDetectCRLF_NewFileOnLinux` — new file gets `\n`
- `TestNormalizeLineEndingsForFile_MixedInput` — mixed `\r\n`+`\n` input normalized
- `TestDetectCRLF_SniffFunction` — pure byte-level detection
