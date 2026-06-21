# Tasks: Fix Cross-OS Line Ending Normalization

## Status: PLANNED

### Phase 1: Planning (DONE)
- [x] OpenSpec proposal + design + tasks
- [x] Issue file with problem statement

### Phase 2: Implementation
- [ ] cursor-byok: Create `internal/agent/lineendings.go` with `normalizeLineEndingsForFile()` helper
- [ ] cursor-byok: Fix Write tool (`tool_exec.go` line 721-746) — normalize before `StreamContent`
- [ ] cursor-byok: Fix StrReplace tool (`tool_exec.go` line 599-654) — always normalize via helper
- [ ] cursor-byok: Fix EditArgs → WriteArgs (`tool_exec.go` line 882-889) — normalize `FileText`
- [ ] cursor-byok: Fix bugbot.go WriteArgs (line 282-290) — same normalization
- [ ] 9router: Fix `open-sse/utils/stream.js` passthrough — strip `\r` after `split("\n")`

### Phase 3: Testing
- [ ] cursor-byok: Create `internal/agent/lineendings_test.go` unit tests
- [ ] Build success: `go build ./...`
- [ ] Test pass: `go test ./internal/agent/...`
- [ ] No regression: existing tests still pass

### Phase 4: Verification & Documentation
- [ ] Evidence file with before/after comparison
- [ ] Git commit: `fix(agent): normalize line endings across OS for Write/StrReplace tools`
