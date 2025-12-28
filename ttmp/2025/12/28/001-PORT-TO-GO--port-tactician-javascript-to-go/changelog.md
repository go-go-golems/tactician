# Changelog

## 2025-12-28

- Initial workspace created


## 2025-12-28

Step 1: Created comprehensive analysis document mapping all JavaScript CLI verbs and flags to Go/Glazed implementation patterns. Created diary for ongoing work tracking.

### Related Files

- /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/ttmp/2025/12/28/001-PORT-TO-GO--port-tactician-javascript-to-go/analysis/01-javascript-to-go-verb-and-flag-mapping.md — Complete mapping of 7 commands and 15 flags


## 2025-12-28

Updated analysis document to use new Glazed API (schema/fields/values/sources) instead of old layers/parameters API. All command pseudocode now uses CommandDefinition, schema.NewSection(), fields.New(), and values.DecodeSectionInto().

### Related Files

- /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/ttmp/2025/12/28/001-PORT-TO-GO--port-tactician-javascript-to-go/analysis/01-javascript-to-go-verb-and-flag-mapping.md — Updated to new API patterns


## 2025-12-28

Added 19 implementation tasks to guide next developer: setup, database layers, schema sections, all 7 commands, helper functions, testing, and optional LLM reranking


## 2025-12-28

Updated command definitions to support batch operations: node show/edit/delete now accept multiple node IDs using TypeStringList. Added batch operations overview section and updated positional arguments handling documentation.


## 2025-12-28

Updated node command tasks to reflect batch operations support using TypeStringList


## 2025-12-28

Added package structure documentation: package github.com/go-go-golems/tactician, directory structure with pkg/commands/{group}/root.go pattern, cmd/tactician/main.go entry point, one file per command, root.go files for explicit group registration


## 2025-12-28

Added file location notes to each command section showing package path and file structure (pkg/commands/{group}/{command}.go and root.go)


## 2025-12-28

Closed (review): Go port complete; see diary + smoke scripts + archive Mermaid exports

