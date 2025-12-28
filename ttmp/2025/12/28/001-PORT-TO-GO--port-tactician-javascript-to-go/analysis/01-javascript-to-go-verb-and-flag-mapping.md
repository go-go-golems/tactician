---
Title: JavaScript to Go Verb and Flag Mapping
Ticket: 001-PORT-TO-GO
Status: active
Topics:
    - port
    - go
    - cli
    - tactician
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Exhaustive mapping of all Tactician JavaScript CLI verbs and flags to Go/Glazed implementation patterns with batch operation support"
LastUpdated: 2025-12-28T13:19:32.706812671-05:00
WhatFor: "Reference document for implementing Go version of Tactician CLI"
WhenToUse: "During implementation to ensure feature parity and correct parameter mapping"
---

# JavaScript to Go Verb and Flag Mapping

## Overview

This document provides an exhaustive analysis of all CLI verbs and flags in the JavaScript version of Tactician, mapping them to Go implementation patterns using Glazed framework's new API (schema/fields/values/sources vocabulary).

### Batch Operations Support

The Go implementation extends the JavaScript version by adding **batch operation support** for commands that benefit from processing multiple items at once:

- **`node show`**: Show multiple nodes in a single command (`node show id1 id2 id3`)
- **`node edit`**: Update status for multiple nodes (`node edit id1 id2 --status complete`)
- **`node delete`**: Delete multiple nodes (`node delete id1 id2 --force`)

Batch operations use `fields.TypeStringList` for positional arguments, allowing multiple values to be passed. All batch-enabled commands remain backward compatible with single-item usage.

Commands that do **not** support batch operations:
- **`apply`**: Applies one tactic at a time (tactics create complex node graphs)
- **`init`**: Initializes a single project
- **`graph`**, **`goals`**, **`history`**, **`search`**: Query/display commands (already operate on collections)

### Required Imports

All command implementations will need these imports:

```go
import (
    "context"
    "fmt"
    
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/values"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
    "github.com/spf13/cobra"
)
```

### Package Structure

The Go implementation follows a standard Go project layout:

**Package Name:** `github.com/go-go-golems/tactician`

**Directory Structure:**
```
tactician/
├── pkg/                          # Core library code
│   ├── db/                       # Database layer
│   │   ├── project.go            # Project database wrapper
│   │   └── tactics.go            # Tactics database wrapper
│   ├── commands/                 # Command implementations
│   │   ├── init/                 # init command group
│   │   │   ├── root.go           # Register init command group
│   │   │   └── init.go           # InitCommand implementation
│   │   ├── node/                 # node command group
│   │   │   ├── root.go           # Register node subcommands group
│   │   │   ├── show.go           # NodeShowCommand
│   │   │   ├── add.go            # NodeAddCommand
│   │   │   ├── edit.go           # NodeEditCommand
│   │   │   └── delete.go         # NodeDeleteCommand
│   │   ├── graph/                # graph command group
│   │   │   ├── root.go           # Register graph command group
│   │   │   └── graph.go          # GraphCommand implementation
│   │   ├── goals/                # goals command group
│   │   │   ├── root.go           # Register goals command group
│   │   │   └── goals.go          # GoalsCommand implementation
│   │   ├── history/              # history command group
│   │   │   ├── root.go           # Register history command group
│   │   │   └── history.go        # HistoryCommand implementation
│   │   ├── search/               # search command group
│   │   │   ├── root.go           # Register search command group
│   │   │   └── search.go         # SearchCommand implementation
│   │   └── apply/                # apply command group
│   │       ├── root.go           # Register apply command group
│   │       └── apply.go          # ApplyCommand implementation
│   └── helpers/                  # Helper functions
│       ├── status.go             # computeNodeStatus, getPendingNodes
│       ├── time.go               # parseRelativeTime
│       └── ranking.go            # rankTactics, checkDependencies
├── cmd/
│   └── tactician/
│       └── main.go               # Main entry point, registers all command groups
├── go.mod
└── go.sum
```

**Key Principles:**

1. **One directory per command group**: Each top-level command (`init`, `node`, `graph`, etc.) has its own directory
2. **One file per command**: Each command implementation lives in its own file
3. **`root.go` in each directory**: Each command group directory contains a `root.go` file that explicitly registers all commands in that group
4. **Core library in `pkg/`**: Reusable code (database wrappers, helpers) lives in `pkg/`
5. **CLI entry point in `cmd/`**: The main executable lives in `cmd/tactician/main.go`

**Example `root.go` pattern:**

```go
// pkg/commands/node/root.go
package node

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/cobra"
)

// RegisterNodeCommands registers all node subcommands to the parent command
func RegisterNodeCommands(parent *cobra.Command) error {
    // Create node command group
    nodeCmd := &cobra.Command{
        Use:   "node",
        Short: "Manage nodes in the project graph",
    }
    
    // Register subcommands
    showCmd, err := NewNodeShowCommand()
    if err != nil {
        return err
    }
    cobraShowCmd, err := cli.BuildCobraCommandFromCommand(showCmd, ...)
    if err != nil {
        return err
    }
    nodeCmd.AddCommand(cobraShowCmd)
    
    // Repeat for add, edit, delete...
    
    parent.AddCommand(nodeCmd)
    return nil
}
```

**Example `main.go` pattern:**

```go
// cmd/tactician/main.go
package main

import (
    "github.com/go-go-golems/tactician/pkg/commands/init"
    "github.com/go-go-golems/tactician/pkg/commands/node"
    "github.com/go-go-golems/tactician/pkg/commands/graph"
    // ... other command groups
    "github.com/spf13/cobra"
)

func main() {
    root := &cobra.Command{
        Use:   "tactician",
        Short: "Decompose software projects into task DAGs using reusable tactics",
    }
    
    // Register all command groups
    init.RegisterInitCommands(root)
    node.RegisterNodeCommands(root)
    graph.RegisterGraphCommands(root)
    // ... register other groups
    
    if err := root.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Command Structure Analysis

### Root Command
- **Name**: `tactician`
- **Description**: "Decompose software projects into task DAGs using reusable tactics"
- **Version**: 1.0.0

### Commands Overview

1. `init` - Initialize a new Tactician project
2. `node` - Manage nodes in the project graph (subcommands: show, add, edit, delete)
3. `graph` - Display the project dependency graph
4. `goals` - List all open (incomplete) goals
5. `history` - View action history and session summary
6. `search` - Search for applicable tactics
7. `apply` - Apply a tactic to create new nodes

---

## Command 1: `init`

### JavaScript Implementation
- **File**: `src/commands/init.js`
- **No flags or arguments**
- **Behavior**: Creates `.tactician/` directory, initializes project and tactics databases, loads default tactics

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/init`
- **Files**: 
  - `pkg/commands/init/init.go` - InitCommand implementation
  - `pkg/commands/init/root.go` - RegisterInitCommands function

### Go Implementation Mapping

```go
// Command structure
type InitCommand struct {
    *cmds.CommandDefinition
}

// Settings (no flags needed, but we can add optional flags)
type InitSettings struct {
    // Future: could add --force, --tactics-dir, etc.
}

// Implementation pseudocode
func (c *InitCommand) Run(ctx context.Context, vals *values.Values) error {
    // 1. Check if .tactician/ exists
    // 2. Create .tactician/ directory
    // 3. Initialize project database (SQLite)
    // 4. Initialize tactics database (SQLite)
    // 5. Load default tactics from YAML
    // 6. Log project_initialized action
    return nil
}

// Constructor
func NewInitCommand() (*InitCommand, error) {
    // Create empty schema (no sections needed for bare command)
    s := schema.NewSchema()
    
    cmdDef := cmds.NewCommandDefinition(
        "init",
        cmds.WithShort("Initialize a new Tactician project"),
        cmds.WithLong("Creates .tactician/ directory and initializes project and tactics databases"),
        cmds.WithSchema(s),
    )
    return &InitCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- None (bare command)

### Database Schema
- Project DB: `nodes`, `edges`, `action_log`, `project` (meta) tables
- Tactics DB: `tactics`, `tactic_dependencies`, `tactic_subtasks` tables

---

## Command 2: `node` (with subcommands)

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/node`
- **Files**:
  - `pkg/commands/node/show.go` - NodeShowCommand implementation
  - `pkg/commands/node/add.go` - NodeAddCommand implementation
  - `pkg/commands/node/edit.go` - NodeEditCommand implementation
  - `pkg/commands/node/delete.go` - NodeDeleteCommand implementation
  - `pkg/commands/node/root.go` - RegisterNodeCommands function

### Subcommand: `node show <id>...`

#### JavaScript Implementation
- **File**: `src/commands/node.js`
- **Arguments**: `<id>` (required, single)
- **No flags**
- **Note**: JavaScript version only supports single node, but Go version will support batch operations

#### Go Implementation Mapping (Batch-Enabled)

```go
type NodeShowCommand struct {
    *cmds.CommandDefinition
}

type NodeShowSettings struct {
    NodeIDs []string `glazed.parameter:"node-ids"` // Supports multiple nodes
}

func (c *NodeShowCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    settings := &NodeShowSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode node show settings: %w", err)
    }
    
    if len(settings.NodeIDs) == 0 {
        return fmt.Errorf("at least one node ID is required")
    }
    
    // Process each node ID
    for _, nodeID := range settings.NodeIDs {
        node := projectDB.GetNode(nodeID)
        if node == nil {
            // Option 1: Skip missing nodes and continue
            // Option 2: Return error immediately
            return fmt.Errorf("node not found: %s", nodeID)
        }
        
        // Output as structured row
        row := types.NewRow(
            types.MRP("id", node.ID),
            types.MRP("type", node.Type),
            types.MRP("output", node.Output),
            types.MRP("status", computeNodeStatus(node)),
            types.MRP("created_by", node.CreatedBy),
            types.MRP("parent_tactic", node.ParentTactic),
            types.MRP("introduced_as", node.IntroducedAs),
            types.MRP("created_at", node.CreatedAt),
            types.MRP("completed_at", node.CompletedAt),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return fmt.Errorf("failed to add row for node %s: %w", nodeID, err)
        }
    }
    
    return nil
}

func NewNodeShowCommand() (*NodeShowCommand, error) {
    // Create glazed section for output formatting
    glazedSection, err := schema.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    
    // Create default section for positional arguments (supports multiple)
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("node-ids", fields.TypeStringList,
                fields.WithHelp("Node ID(s) to show (supports multiple: node show id1 id2 id3)"),
                fields.WithRequired(true),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(glazedSection, defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "show",
        cmds.WithShort("Show details for one or more nodes"),
        cmds.WithLong("Show details for nodes. Supports batch operations: 'node show id1 id2 id3'"),
        cmds.WithSchema(s),
    )
    return &NodeShowCommand{CommandDefinition: cmdDef}, nil
}
```

**Usage Examples:**
```bash
# Single node (backward compatible)
tactician node show node-123

# Multiple nodes (batch operation)
tactician node show node-123 node-456 node-789

# Output can be formatted with Glazed flags
tactician node show node-123 node-456 --output json
tactician node show node-123 node-456 --fields id,status,output
```

### Subcommand: `node add <id> <output>`

#### JavaScript Implementation
- **Arguments**: `<id>` (required), `<output>` (required)
- **Flags**:
  - `--type <type>` - Node type (default: 'project_artifact')
  - `--status <status>` - Initial status (default: 'pending')

#### Go Implementation Mapping

```go
type NodeAddCommand struct {
    *cmds.CommandDefinition
}

type NodeAddSettings struct {
    NodeID  string `glazed.parameter:"node-id"`
    Output  string `glazed.parameter:"output"`
    Type    string `glazed.parameter:"type"`
    Status  string `glazed.parameter:"status"`
}

func (c *NodeAddCommand) Run(ctx context.Context, vals *values.Values) error {
    settings := &NodeAddSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode node add settings: %w", err)
    }
    
    // Validate node doesn't exist
    if projectDB.GetNode(settings.NodeID) != nil {
        return fmt.Errorf("node already exists: %s", settings.NodeID)
    }
    
    node := &Node{
        ID:        settings.NodeID,
        Type:      settings.Type,
        Output:    settings.Output,
        Status:    settings.Status,
        CreatedAt: time.Now(),
    }
    
    projectDB.AddNode(node)
    projectDB.LogAction("node_created", fmt.Sprintf("Created node: %s", settings.NodeID), settings.NodeID)
    return nil
}

func NewNodeAddCommand() (*NodeAddCommand, error) {
    // Create default section for positional arguments and flags
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("node-id", fields.TypeString,
                fields.WithHelp("Node ID"),
                fields.WithRequired(true),
            ),
            fields.New("output", fields.TypeString,
                fields.WithHelp("Node output artifact"),
                fields.WithRequired(true),
            ),
        ),
        schema.WithFields(
            fields.New("type", fields.TypeString,
                fields.WithHelp("Node type"),
                fields.WithDefault("project_artifact"),
            ),
            fields.New("status", fields.TypeChoice,
                fields.WithHelp("Initial status"),
                fields.WithChoices("pending", "complete"),
                fields.WithDefault("pending"),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "add",
        cmds.WithShort("Add a new node"),
        cmds.WithSchema(s),
    )
    return &NodeAddCommand{CommandDefinition: cmdDef}, nil
}
```

### Subcommand: `node edit <id>...`

#### JavaScript Implementation
- **Arguments**: `<id>` (required, single)
- **Flags**:
  - `--status <status>` - Update status (pending, complete)
- **Note**: JavaScript version only supports single node, but Go version will support batch operations

#### Go Implementation Mapping (Batch-Enabled)

```go
type NodeEditCommand struct {
    *cmds.CommandDefinition
}

type NodeEditSettings struct {
    NodeIDs []string `glazed.parameter:"node-ids"` // Supports multiple nodes
    Status  string   `glazed.parameter:"status"`   // Applied to all nodes
}

func (c *NodeEditCommand) Run(ctx context.Context, vals *values.Values) error {
    settings := &NodeEditSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode node edit settings: %w", err)
    }
    
    if len(settings.NodeIDs) == 0 {
        return fmt.Errorf("at least one node ID is required")
    }
    
    if settings.Status == "" {
        return fmt.Errorf("--status flag is required")
    }
    
    completedAt := time.Time{}
    if settings.Status == "complete" {
        completedAt = time.Now()
    }
    
    // Update all specified nodes
    for _, nodeID := range settings.NodeIDs {
        node := projectDB.GetNode(nodeID)
        if node == nil {
            return fmt.Errorf("node not found: %s", nodeID)
        }
        
        projectDB.UpdateNodeStatus(nodeID, settings.Status, completedAt)
        
        action := "node_updated"
        if settings.Status == "complete" {
            action = "node_completed"
        }
        projectDB.LogAction(action, fmt.Sprintf("Updated %s status to %s", nodeID, settings.Status), nodeID)
    }
    
    return nil
}

func NewNodeEditCommand() (*NodeEditCommand, error) {
    // Create default section
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("node-ids", fields.TypeStringList,
                fields.WithHelp("Node ID(s) to edit (supports multiple: node edit id1 id2 --status complete)"),
                fields.WithRequired(true),
            ),
        ),
        schema.WithFields(
            fields.New("status", fields.TypeChoice,
                fields.WithHelp("Update status (applied to all specified nodes)"),
                fields.WithChoices("pending", "complete"),
                fields.WithRequired(true),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "edit",
        cmds.WithShort("Edit one or more nodes"),
        cmds.WithLong("Update status for nodes. Supports batch operations: 'node edit id1 id2 --status complete'"),
        cmds.WithSchema(s),
    )
    return &NodeEditCommand{CommandDefinition: cmdDef}, nil
}
```

**Usage Examples:**
```bash
# Single node (backward compatible)
tactician node edit node-123 --status complete

# Multiple nodes (batch operation)
tactician node edit node-123 node-456 node-789 --status complete
```

### Subcommand: `node delete <id>...`

#### JavaScript Implementation
- **Arguments**: `<id>` (required, single)
- **Flags**:
  - `--force` - Force delete even if it blocks other nodes
- **Note**: JavaScript version only supports single node, but Go version will support batch operations

#### Go Implementation Mapping (Batch-Enabled)

```go
type NodeDeleteCommand struct {
    *cmds.CommandDefinition
}

type NodeDeleteSettings struct {
    NodeIDs []string `glazed.parameter:"node-ids"` // Supports multiple nodes
    Force   bool     `glazed.parameter:"force"`
}

func (c *NodeDeleteCommand) Run(ctx context.Context, vals *values.Values) error {
    settings := &NodeDeleteSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode node delete settings: %w", err)
    }
    
    if len(settings.NodeIDs) == 0 {
        return fmt.Errorf("at least one node ID is required")
    }
    
    // Check all nodes exist and validate deletion constraints
    var nodesToDelete []string
    for _, nodeID := range settings.NodeIDs {
        node := projectDB.GetNode(nodeID)
        if node == nil {
            return fmt.Errorf("node not found: %s", nodeID)
        }
        
        blocks := projectDB.GetBlockedBy(nodeID)
        if len(blocks) > 0 && !settings.Force {
            return fmt.Errorf("cannot delete %s: it blocks %d node(s). Use --force to delete anyway", nodeID, len(blocks))
        }
        
        nodesToDelete = append(nodesToDelete, nodeID)
    }
    
    // Delete all nodes
    for _, nodeID := range nodesToDelete {
        projectDB.DeleteNode(nodeID)
        projectDB.LogAction("node_deleted", fmt.Sprintf("Deleted node: %s", nodeID), nodeID)
    }
    
    return nil
}

func NewNodeDeleteCommand() (*NodeDeleteCommand, error) {
    // Create default section
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("node-ids", fields.TypeStringList,
                fields.WithHelp("Node ID(s) to delete (supports multiple: node delete id1 id2 --force)"),
                fields.WithRequired(true),
            ),
        ),
        schema.WithFields(
            fields.New("force", fields.TypeBool,
                fields.WithHelp("Force delete even if it blocks other nodes (applies to all specified nodes)"),
                fields.WithDefault(false),
                fields.WithShortFlag("f"),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "delete",
        cmds.WithShort("Delete one or more nodes"),
        cmds.WithLong("Delete nodes. Supports batch operations: 'node delete id1 id2 --force'"),
        cmds.WithSchema(s),
    )
    return &NodeDeleteCommand{CommandDefinition: cmdDef}, nil
}
```

**Usage Examples:**
```bash
# Single node (backward compatible)
tactician node delete node-123

# Multiple nodes (batch operation)
tactician node delete node-123 node-456 node-789

# With force flag
tactician node delete node-123 node-456 --force
```

---

## Command 3: `graph [goal-id]`

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/graph`
- **Files**:
  - `pkg/commands/graph/graph.go` - GraphCommand implementation
  - `pkg/commands/graph/root.go` - RegisterGraphCommands function

### JavaScript Implementation
- **Arguments**: `[goal-id]` (optional)
- **Flags**:
  - `--mermaid` - Output as Mermaid diagram

### Go Implementation Mapping

```go
type GraphCommand struct {
    *cmds.CommandDefinition
}

type GraphSettings struct {
    GoalID  string `glazed.parameter:"goal-id"`
    Mermaid bool   `glazed.parameter:"mermaid"`
}

func (c *GraphCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    settings := &GraphSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode graph settings: %w", err)
    }
    
    if settings.Mermaid {
        // Generate Mermaid output
        return c.generateMermaidGraph(ctx, gp, settings.GoalID)
    }
    
    // Generate structured tree output
    return c.generateTreeOutput(ctx, gp, settings.GoalID)
}

func NewGraphCommand() (*GraphCommand, error) {
    // Create glazed section for output formatting
    glazedSection, err := schema.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    
    // Create default section for optional positional argument and flag
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("goal-id", fields.TypeString,
                fields.WithHelp("Start graph from specific goal node"),
            ),
        ),
        schema.WithFields(
            fields.New("mermaid", fields.TypeBool,
                fields.WithHelp("Output as Mermaid diagram"),
                fields.WithDefault(false),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(glazedSection, defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "graph",
        cmds.WithShort("Display the project dependency graph"),
        cmds.WithSchema(s),
    )
    return &GraphCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- Glazed section (for output formatting)
- Default section (for goal-id and mermaid flag)

---

## Command 4: `goals`

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/goals`
- **Files**:
  - `pkg/commands/goals/goals.go` - GoalsCommand implementation
  - `pkg/commands/goals/root.go` - RegisterGoalsCommands function

### JavaScript Implementation
- **Flags**:
  - `--mermaid` - Output as Mermaid diagram

### Go Implementation Mapping

```go
type GoalsCommand struct {
    *cmds.CommandDefinition
}

type GoalsSettings struct {
    Mermaid bool `glazed.parameter:"mermaid"`
}

func (c *GoalsCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    settings := &GoalsSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode goals settings: %w", err)
    }
    
    pendingNodes := getPendingNodes(projectDB)
    
    if settings.Mermaid {
        return c.generateMermaidGoals(ctx, gp, pendingNodes)
    }
    
    // Output as structured rows
    for _, node := range pendingNodes {
        actualStatus := computeNodeStatus(node, projectDB)
        deps := projectDB.GetDependencies(node.ID)
        blocks := projectDB.GetBlockedBy(node.ID)
        
        row := types.NewRow(
            types.MRP("id", node.ID),
            types.MRP("output", node.Output),
            types.MRP("status", actualStatus),
            types.MRP("dependencies", deps),
            types.MRP("blocks", blocks),
            types.MRP("parent_tactic", node.ParentTactic),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

func NewGoalsCommand() (*GoalsCommand, error) {
    // Create glazed section for output formatting
    glazedSection, err := schema.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    
    // Create default section for mermaid flag
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithFields(
            fields.New("mermaid", fields.TypeBool,
                fields.WithHelp("Output as Mermaid diagram"),
                fields.WithDefault(false),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(glazedSection, defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "goals",
        cmds.WithShort("List all open (incomplete) goals"),
        cmds.WithSchema(s),
    )
    return &GoalsCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- Glazed section (for output formatting)
- Default section (for mermaid flag)

---

## Command 5: `history`

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/history`
- **Files**:
  - `pkg/commands/history/history.go` - HistoryCommand implementation
  - `pkg/commands/history/root.go` - RegisterHistoryCommands function

### JavaScript Implementation
- **Flags**:
  - `-l, --limit <number>` - Limit number of entries (default: unlimited)
  - `-s, --since <time>` - Show actions since (e.g., 1h, 2d, 30m)
  - `--summary` - Show session summary instead of detailed log

### Go Implementation Mapping

```go
type HistoryCommand struct {
    *cmds.CommandDefinition
}

type HistorySettings struct {
    Limit   int    `glazed.parameter:"limit"`
    Since   string `glazed.parameter:"since"`
    Summary bool   `glazed.parameter:"summary"`
}

func (c *HistoryCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    settings := &HistorySettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode history settings: %w", err)
    }
    
    var sinceTime time.Time
    if settings.Since != "" {
        // Parse relative time (1h, 2d, 30m)
        sinceTime = parseRelativeTime(settings.Since)
    }
    
    if settings.Summary {
        return c.outputSummary(ctx, gp, sinceTime)
    }
    
    logs := projectDB.GetActionLog(settings.Limit, sinceTime)
    
    for _, log := range logs {
        row := types.NewRow(
            types.MRP("timestamp", log.Timestamp),
            types.MRP("action", log.Action),
            types.MRP("details", log.Details),
            types.MRP("node_id", log.NodeID),
            types.MRP("tactic_id", log.TacticID),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

func NewHistoryCommand() (*HistoryCommand, error) {
    // Create glazed section for output formatting
    glazedSection, err := schema.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    
    // Create default section for flags
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithFields(
            fields.New("limit", fields.TypeInteger,
                fields.WithHelp("Limit number of entries"),
                fields.WithShortFlag("l"),
            ),
            fields.New("since", fields.TypeString,
                fields.WithHelp("Show actions since (e.g., 1h, 2d, 30m)"),
                fields.WithShortFlag("s"),
            ),
            fields.New("summary", fields.TypeBool,
                fields.WithHelp("Show session summary instead of detailed log"),
                fields.WithDefault(false),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(glazedSection, defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "history",
        cmds.WithShort("View action history and session summary"),
        cmds.WithSchema(s),
    )
    return &HistoryCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- Glazed section (for output formatting)
- Default section (for limit, since, summary flags)

---

## Command 6: `search [query]`

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/search`
- **Files**:
  - `pkg/commands/search/search.go` - SearchCommand implementation
  - `pkg/commands/search/root.go` - RegisterSearchCommands function

### JavaScript Implementation
- **Arguments**: `[query]` (optional)
- **Flags**:
  - `--ready` - Show only ready tactics (all dependencies satisfied)
  - `--type <type>` - Filter by tactic type
  - `--tags <tags>` - Filter by tags (comma-separated)
  - `--goals <goals>` - Align with specific goal nodes (comma-separated)
  - `--llm-rerank` - Use LLM to semantically rerank results
  - `-l, --limit <number>` - Limit number of results (default: 20)
  - `-v, --verbose` - Show detailed scoring information

### Go Implementation Mapping

```go
type SearchCommand struct {
    *cmds.CommandDefinition
}

type SearchSettings struct {
    Query      string `glazed.parameter:"query"`
    Ready      bool   `glazed.parameter:"ready"`
    Type       string `glazed.parameter:"type"`
    Tags       string `glazed.parameter:"tags"`
    Goals      string `glazed.parameter:"goals"`
    LLMRerank  bool   `glazed.parameter:"llm-rerank"`
    Limit      int    `glazed.parameter:"limit"`
    Verbose    bool   `glazed.parameter:"verbose"`
}

func (c *SearchCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    settings := &SearchSettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode search settings: %w", err)
    }
    
    tactics := tacticsDB.GetAllTactics()
    
    // Apply filters
    if settings.Type != "" {
        tactics = filterByType(tactics, settings.Type)
    }
    
    if settings.Tags != "" {
        tagList := strings.Split(settings.Tags, ",")
        tactics = filterByTags(tactics, tagList)
    }
    
    // Parse keywords from query
    keywords := parseKeywords(settings.Query)
    
    // Parse goal IDs
    goalIds := parseGoalIds(settings.Goals)
    
    // Rank tactics
    ranked := rankTactics(tactics, projectDB, keywords, goalIds)
    
    // Filter by ready if requested
    if settings.Ready {
        ranked = filterReady(ranked)
    }
    
    // Apply LLM reranking if requested
    if settings.LLMRerank {
        ranked = await llmReranker.Rerank(settings.Query, ranked, projectDB)
    }
    
    // Limit results
    limit := settings.Limit
    if limit == 0 {
        limit = 20
    }
    results := ranked[:limit]
    
    // Output as structured rows
    for _, result := range results {
        tactic := result.Tactic
        depStatus := result.DepStatus
        
        row := types.NewRow(
            types.MRP("id", tactic.ID),
            types.MRP("type", tactic.Type),
            types.MRP("output", tactic.Output),
            types.MRP("ready", depStatus.Ready),
            types.MRP("satisfied", depStatus.Satisfied),
            types.MRP("missing", depStatus.Missing),
            types.MRP("can_introduce", depStatus.CanIntroduce),
            types.MRP("tags", tactic.Tags),
            types.MRP("description", tactic.Description),
        )
        
        if settings.Verbose {
            row.Set("scores", result.Scores)
        }
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

func NewSearchCommand() (*SearchCommand, error) {
    // Create glazed section for output formatting
    glazedSection, err := schema.NewGlazedSchema()
    if err != nil {
        return nil, err
    }
    
    // Create default section for optional query and flags
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("query", fields.TypeString,
                fields.WithHelp("Search query (keywords)"),
            ),
        ),
        schema.WithFields(
            fields.New("ready", fields.TypeBool,
                fields.WithHelp("Show only ready tactics (all dependencies satisfied)"),
                fields.WithDefault(false),
            ),
            fields.New("type", fields.TypeString,
                fields.WithHelp("Filter by tactic type"),
            ),
            fields.New("tags", fields.TypeString,
                fields.WithHelp("Filter by tags (comma-separated)"),
            ),
            fields.New("goals", fields.TypeString,
                fields.WithHelp("Align with specific goal nodes (comma-separated)"),
            ),
            fields.New("llm-rerank", fields.TypeBool,
                fields.WithHelp("Use LLM to semantically rerank results"),
                fields.WithDefault(false),
            ),
            fields.New("limit", fields.TypeInteger,
                fields.WithHelp("Limit number of results"),
                fields.WithDefault(20),
                fields.WithShortFlag("l"),
            ),
            fields.New("verbose", fields.TypeBool,
                fields.WithHelp("Show detailed scoring information"),
                fields.WithDefault(false),
                fields.WithShortFlag("v"),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(glazedSection, defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "search",
        cmds.WithShort("Search for applicable tactics"),
        cmds.WithSchema(s),
    )
    return &SearchCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- Glazed section (for output formatting)
- Default section (for query, ready, type, tags, goals, llm-rerank, limit, verbose)
- Potentially: LLM section (for reranking configuration)

---

## Command 7: `apply <tactic-id>`

### Go Implementation Location
- **Package**: `github.com/go-go-golems/tactician/pkg/commands/apply`
- **Files**:
  - `pkg/commands/apply/apply.go` - ApplyCommand implementation
  - `pkg/commands/apply/root.go` - RegisterApplyCommands function

### JavaScript Implementation
- **Arguments**: `<tactic-id>` (required)
- **Flags**:
  - `-y, --yes` - Skip confirmation prompt
  - `-f, --force` - Apply even if dependencies are missing

### Go Implementation Mapping

```go
type ApplyCommand struct {
    *cmds.CommandDefinition
}

type ApplySettings struct {
    TacticID string `glazed.parameter:"tactic-id"`
    Yes      bool   `glazed.parameter:"yes"`
    Force    bool   `glazed.parameter:"force"`
}

func (c *ApplyCommand) Run(ctx context.Context, vals *values.Values) error {
    settings := &ApplySettings{}
    if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to decode apply settings: %w", err)
    }
    
    tactic := tacticsDB.GetTactic(settings.TacticID)
    if tactic == nil {
        return fmt.Errorf("tactic not found: %s", settings.TacticID)
    }
    
    // Check dependencies
    depStatus := checkDependencies(tactic, projectDB)
    
    // Show dependency status
    if len(depStatus.Missing) > 0 && !settings.Force {
        return fmt.Errorf("cannot apply tactic: missing required dependencies. Use --force to apply anyway")
    }
    
    // Determine nodes to create
    nodesToCreate := buildNodesToCreate(tactic, depStatus)
    
    // Confirm unless --yes
    if !settings.Yes {
        if !confirmApply(nodesToCreate) {
            return nil // User cancelled
        }
    }
    
    // Create nodes
    timestamp := time.Now()
    for _, node := range nodesToCreate {
        node.CreatedAt = timestamp
        projectDB.AddNode(node)
    }
    
    // Create edges for dependencies
    createEdges(tactic, nodesToCreate, depStatus, projectDB)
    
    // Log action
    projectDB.LogAction("tactic_applied", fmt.Sprintf("Applied tactic: %s", settings.TacticID), nil, settings.TacticID)
    
    return nil
}

func NewApplyCommand() (*ApplyCommand, error) {
    // Create default section for positional argument and flags
    defaultSection, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithDescription("Default parameters"),
        schema.WithArguments(
            fields.New("tactic-id", fields.TypeString,
                fields.WithHelp("Tactic ID to apply"),
                fields.WithRequired(true),
            ),
        ),
        schema.WithFields(
            fields.New("yes", fields.TypeBool,
                fields.WithHelp("Skip confirmation prompt"),
                fields.WithDefault(false),
                fields.WithShortFlag("y"),
            ),
            fields.New("force", fields.TypeBool,
                fields.WithHelp("Apply even if dependencies are missing"),
                fields.WithDefault(false),
                fields.WithShortFlag("f"),
            ),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create schema
    s := schema.NewSchema(
        schema.WithSections(defaultSection),
    )
    
    cmdDef := cmds.NewCommandDefinition(
        "apply",
        cmds.WithShort("Apply a tactic to create new nodes"),
        cmds.WithSchema(s),
    )
    return &ApplyCommand{CommandDefinition: cmdDef}, nil
}
```

### Schema Sections Required
- Default section (for tactic-id, yes, force)
- None (bare command, but could use Glazed section for structured output of what will be created)

---

## Common Patterns and Schema Sections

### Identity Schema Sections

Based on the custom layer tutorial pattern, we should create reusable schema sections for:

1. **Project Section** - Common project database access
   - `project-db-path` (default: `.tactician/project.db`)
   - `tactics-db-path` (default: `.tactician/tactics.db`)

2. **Output Format Section** - Common output options
   - `mermaid` - Output as Mermaid diagram
   - Already covered by Glazed section for most commands

3. **Database Section** - Database connection and configuration
   - Could include connection pooling, WAL mode, etc.

### Pseudocode for Identity Schema Sections

```go
// Project Section
const ProjectSlug = "project"

type ProjectSettings struct {
    ProjectDBPath string `glazed.parameter:"project-db-path"`
    TacticsDBPath string `glazed.parameter:"tactics-db-path"`
}

func NewProjectSection() (*schema.Section, error) {
    return schema.NewSection(
        ProjectSlug,
        "Project",
        schema.WithPrefix("project-"),
        schema.WithDescription("Project Database Configuration"),
        schema.WithFields(
            fields.New("db-path", fields.TypeString,
                fields.WithHelp("Path to project database"),
                fields.WithDefault(".tactician/project.db"),
            ),
            fields.New("tactics-db-path", fields.TypeString,
                fields.WithHelp("Path to tactics database"),
                fields.WithDefault(".tactician/tactics.db"),
            ),
        ),
    )
}

// Usage in commands
func (c *SomeCommand) Run(ctx context.Context, vals *values.Values) error {
    projectSettings := &ProjectSettings{}
    if err := values.DecodeSectionInto(vals, ProjectSlug, projectSettings); err != nil {
        return fmt.Errorf("failed to decode project settings: %w", err)
    }
    
    projectDB := NewProjectDB(projectSettings.ProjectDBPath)
    // ...
}
```

---

## Database Schema Mapping

### Project Database Tables

1. **project** (meta)
   - `key TEXT PRIMARY KEY`
   - `value TEXT`

2. **nodes**
   - `id TEXT PRIMARY KEY`
   - `type TEXT NOT NULL`
   - `output TEXT NOT NULL`
   - `status TEXT NOT NULL DEFAULT 'pending'`
   - `created_by TEXT`
   - `created_at TEXT NOT NULL`
   - `completed_at TEXT`
   - `parent_tactic TEXT`
   - `introduced_as TEXT`
   - `data TEXT` (JSON)

3. **edges**
   - `id INTEGER PRIMARY KEY AUTOINCREMENT`
   - `source_node_id TEXT NOT NULL`
   - `target_node_id TEXT NOT NULL`
   - Foreign keys with CASCADE delete

4. **action_log**
   - `id INTEGER PRIMARY KEY AUTOINCREMENT`
   - `timestamp TEXT NOT NULL`
   - `action TEXT NOT NULL`
   - `details TEXT`
   - `node_id TEXT`
   - `tactic_id TEXT`

### Tactics Database Tables

1. **tactics**
   - `id TEXT PRIMARY KEY`
   - `type TEXT NOT NULL`
   - `output TEXT NOT NULL`
   - `description TEXT`
   - `tags TEXT` (comma-separated)
   - `data TEXT` (JSON)

2. **tactic_dependencies**
   - `id INTEGER PRIMARY KEY AUTOINCREMENT`
   - `tactic_id TEXT NOT NULL`
   - `dependency_type TEXT NOT NULL` ('match' or 'premise')
   - `artifact_type TEXT NOT NULL`

3. **tactic_subtasks**
   - `id INTEGER PRIMARY KEY AUTOINCREMENT`
   - `tactic_id TEXT NOT NULL`
   - `subtask_id TEXT NOT NULL`
   - `output TEXT NOT NULL`
   - `type TEXT NOT NULL`
   - `depends_on TEXT` (JSON array)
   - `data TEXT` (JSON)

---

## Summary of All Flags

### By Command

1. **init**: No flags
2. **node show**: No flags
3. **node add**: `--type`, `--status`
4. **node edit**: `--status`
5. **node delete**: `--force`
6. **graph**: `--mermaid`, `[goal-id]` (positional)
7. **goals**: `--mermaid`
8. **history**: `-l/--limit`, `-s/--since`, `--summary`
9. **search**: `--ready`, `--type`, `--tags`, `--goals`, `--llm-rerank`, `-l/--limit`, `-v/--verbose`, `[query]` (positional)
10. **apply**: `-y/--yes`, `-f/--force`, `<tactic-id>` (positional)

### Total Unique Flags: 15

1. `--type` (string)
2. `--status` (choice: pending, complete)
3. `--force` (bool)
4. `--mermaid` (bool)
5. `--limit` (integer)
6. `--since` (string, relative time)
7. `--summary` (bool)
8. `--ready` (bool)
9. `--tags` (string, comma-separated)
10. `--goals` (string, comma-separated)
11. `--llm-rerank` (bool)
12. `--verbose` (bool)
13. `--yes` (bool)
14. `--query` (string, positional in search)
15. `--goal-id` (string, positional in graph)

---

## Implementation Notes

### Positional Arguments Handling

The new API uses `schema.WithArguments()` to define positional arguments in the default section. For commands with positional args:

**Single-value arguments:**
- `node add <id> <output>` - Use `schema.WithArguments()` with two required string fields
- `search [query]` - Optional string argument (not required)
- `graph [goal-id]` - Optional string argument (not required)
- `apply <tactic-id>` - Required string argument

**Batch operations (multiple values):**
- `node show <id>...` - Use `fields.TypeStringList` to accept multiple node IDs
- `node edit <id>...` - Use `fields.TypeStringList` to accept multiple node IDs
- `node delete <id>...` - Use `fields.TypeStringList` to accept multiple node IDs

**Example batch argument definition:**
```go
schema.WithArguments(
    fields.New("node-ids", fields.TypeStringList,
        fields.WithHelp("Node ID(s) to process (supports multiple)"),
        fields.WithRequired(true),
    ),
)
```

**Decoding batch arguments:**
```go
type Settings struct {
    NodeIDs []string `glazed.parameter:"node-ids"` // Automatically populated as slice
}

// In command implementation:
for _, nodeID := range settings.NodeIDs {
    // Process each node
}
```

### Value Decoding Pattern

All commands should use the new value decoding pattern:
1. Create settings struct with `glazed.parameter` tags
2. Use `values.DecodeSectionInto(vals, sectionSlug, &settings)` to populate struct
3. Handle errors with context

### Database Access Pattern

All commands that need database access should:
1. Decode settings from values using `values.DecodeSectionInto()`
2. Open database connection
3. Perform operations
4. Close database connection (defer pattern)

### Error Handling

- Use `github.com/pkg/errors` for error wrapping
- Return descriptive errors with context
- Use structured logging for debugging

### Output Formatting

- Commands that output structured data use `RunIntoGlazeProcessor` with `*values.Values`
- Commands that perform actions use `Run` (bare command) with `*values.Values`
- Mermaid output can be handled via conditional logic or separate formatter

### CLI Integration

Use `cli.BuildCobraCommandFromCommand()` instead of `cli.BuildCobraCommand()`:

```go
cobraCmd, err := cli.BuildCobraCommandFromCommand(
    cmd,
    cli.WithParserConfig(cli.CobraParserConfig{
        AppName: "tactician", // Enables env var parsing: TACTICIAN_<SECTION>_<FIELD>
    }),
)
```

---

## Next Steps

1. **Create Go module structure**:
   - Initialize module: `go mod init github.com/go-go-golems/tactician`
   - Create directory structure: `pkg/`, `cmd/tactician/`, `pkg/commands/{init,node,graph,goals,history,search,apply}/`
   - Create `root.go` files in each command group directory
   - Create `cmd/tactician/main.go` as entry point

2. Implement database layer (SQLite with better-sqlite3 equivalent)
3. Implement identity schema sections (Project, Database) using `schema.NewSection()`
4. Implement commands in order of dependency using new API:
   - `init` (no dependencies) - Use `cmds.NewCommandDefinition()` with empty schema
   - `node` subcommands (depend on init) - Use `schema.WithArguments()` for positional args
   - `graph`, `goals` (depend on project DB) - Use `schema.NewGlazedSchema()` for output formatting
   - `search` (depends on tactics DB) - Use `values.DecodeSectionInto()` for settings
   - `apply` (depends on both DBs) - Use `Run()` with `*values.Values`
   - `history` (depends on project DB) - Use `RunIntoGlazeProcessor()` with `*values.Values`
5. Integrate with Cobra using `cli.BuildCobraCommandFromCommand()`
6. Add tests for each command
7. Add integration tests for full workflows

## API Migration Summary

**Old API → New API:**
- `cmds.CommandDescription` → `cmds.CommandDefinition`
- `cmds.NewCommandDescription()` → `cmds.NewCommandDefinition()`
- `layers.ParameterLayer` → `schema.Section`
- `layers.NewParameterLayer()` → `schema.NewSection()`
- `parameters.NewParameterDefinition()` → `fields.New()`
- `layers.ParameterLayers` → `schema.Schema`
- `layers.ParsedLayers` → `values.Values`
- `parsedLayers.InitializeStruct()` → `values.DecodeSectionInto()`
- `settings.NewGlazedParameterLayers()` → `schema.NewGlazedSchema()`
- `cmds.WithFlags()` → `schema.WithFields()` or `schema.WithArguments()`
- `cmds.WithLayersList()` → `schema.WithSections()` + `cmds.WithSchema()`
- `cli.BuildCobraCommand()` → `cli.BuildCobraCommandFromCommand()`
