# Proposal: Fix Cross-OS Line Ending Normalization in MITM Tool Pipeline

## Status: PROPOSED

## Problem Statement

When using the cursor-byok MITM bridge with 9router as the upstream provider, files written through the Write and StrReplace tools end up with double line endings (`\r\n\r\n` instead of `\r\n` on Windows, or inconsistent mixed endings across OS).

## Root Cause Analysis

The MITM data flow for file writes:

```
Cursor IDE → cursor-byok (MITM) → 9router → LLM Provider
                                           ↓
Cursor IDE ← cursor-byok (protobuf) ← 9router (SSE) ← LLM
```

### cursor-byok Issues (PRIMARY — causes the file double-spacing)

1. **Write tool** (`tool_exec.go` line 721-746): ZERO line ending handling. Whatever the model generates is passed directly to Cursor IDE via protobuf `EditArgs.StreamContent`.

2. **StrReplace tool** (`tool_exec.go` line 599-654): Has `\r\n` normalization but ONLY in a fallback path (when literal match fails). The primary literal-match path has no normalization.

3. **EditArgs → WriteArgs conversion** (`tool_exec.go` line 882-889) and **bugbot.go** (line 282-290): Pass `FileText` to Cursor IDE without normalization.

### 9router Issue (SECONDARY — affects SSE wire framing, not file content)

4. **Passthrough mode** (`open-sse/utils/stream.js`): Buffer splits on `\n` but leaves `\r` on each line when upstream sends `\r\n`. This causes `\r\n\r\n` on the SSE wire but NOT in file content (JSON parsing strips `\r` from parsed values).

## Proposed Solution

### cursor-byok (primary fix)
- Add `normalizeLineEndingsForFile(path, content)` helper that:
  1. Always collapses `\r\n` → `\n` first (clean baseline)
  2. Detects target file's existing convention by sniffing first `\r\n` vs `\n`
  3. Falls back to OS default (`\r\n` on Windows, `\n` on Linux/macOS)
  4. Applies the detected convention to the content
- Apply to Write, StrReplace, EditArgs→WriteArgs, and bugbot WriteArgs paths

### 9router (secondary fix)
- Strip `\r` after `buffer.split("\n")` in passthrough mode

## Scope
- **In scope**: File write line ending normalization, SSE wire framing
- **Out of scope**: Read tool output normalization (Cursor IDE handles this), model prompt text, conversation history

## Dependencies
- `runtime.GOOS` for OS default detection
- `os.ReadFile` for existing file sniffing

## Risks
- Binary files: `normalizeLineEndingsForFile` should be safe since models don't generate binary content, but a safety check on empty/short content is included
- Performance: Single pass `\r\n` → `\n` + conditional `\n` → `\r\n` is O(n), negligible for file-sized content
