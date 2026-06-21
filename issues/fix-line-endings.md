# Issue: Double-spacing / \r\n\r\n in files written through MITM bridge

**Status**: OPEN
**Priority**: HIGH
**Component**: cursor-byok (primary), 9router (secondary)
**Evidence**: [fix-line-endings.md](./evidence/fix-line-endings.md)

## Symptoms

Files edited through the Cursor agent (Write / StrReplace tools) via the cursor-byok MITM bridge + 9router upstream end up with:
- Double blank lines between every line (`\r\n\r\n`)
- Inconsistent line endings (mixed `\r\n` and `\n` in the same file)
- Affects Windows most visibly, but Linux/macOS files can get wrong endings too

## Affected Paths

### cursor-byok (primary cause)
- `internal/agent/tool_exec.go` — Write tool (line 721-746): NO normalization
- `internal/agent/tool_exec.go` — StrReplace tool (line 599-654): normalization only in fallback path
- `internal/agent/tool_exec.go` — EditArgs→WriteArgs (line 882-889): NO normalization
- `internal/agent/bugbot.go` — WriteArgs (line 282-290): NO normalization

### 9router (secondary cause)
- `open-sse/utils/stream.js` — passthrough mode: `\r` not stripped after `split("\n")`

## Root Cause

The LLM model generates text with `\n` line endings. cursor-byok passes it through to Cursor IDE's protobuf `EditArgs.StreamContent` without normalization. Cursor IDE writes the file. On Windows, the file system / git config may add `\r\n`, and the next read shows mixed endings. StrReplace's fallback normalization is inconsistent (only triggers on literal match failure).

## Proposed Fix

See [openspec/changes/fix-line-endings/proposal.md](../openspec/changes/fix-line-endings/proposal.md)

## Acceptance Criteria

- [ ] Write tool normalizes to target file's convention (detect existing, fall back to OS default)
- [ ] StrReplace ALWAYS normalizes (not just in fallback path)
- [ ] EditArgs→WriteArgs conversion normalizes FileText
- [ ] Bugbot WriteArgs path normalizes FileText
- [ ] 9router passthrough mode strips `\r` from split lines
- [ ] Unit tests cover Windows CRLF, Linux LF, macOS LF, new file, mixed input
- [ ] Build + test pass with no regression
