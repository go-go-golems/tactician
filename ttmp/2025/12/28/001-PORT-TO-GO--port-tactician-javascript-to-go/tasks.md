# Tasks

## TODO

- [ ] Setup Go module structure: create go.mod with package github.com/go-go-golems/tactician, add dependencies (glazed, cobra, sqlite driver), create directory structure: pkg/commands/{init,node,graph,goals,history,search,apply}/ with root.go in each, create cmd/tactician/main.go
- [ ] Implement database layer: create SQLite wrapper for project database (nodes, edges, action_log, project meta tables) matching JavaScript API
- [ ] Implement database layer: create SQLite wrapper for tactics database (tactics, tactic_dependencies, tactic_subtasks tables) matching JavaScript API
- [ ] Create Project schema section: implement reusable schema section for project database paths (project-db-path, tactics-db-path) using schema.NewSection()
- [ ] Implement init command: create InitCommand using CommandDefinition, initialize .tactician/ directory, create databases, load default tactics from YAML
- [ ] Implement node show command: create NodeShowCommand with RunIntoGlazeProcessor, use schema.WithArguments() with TypeStringList for node-ids (supports batch: node show id1 id2 id3), decode with values.DecodeSectionInto(), loop through all node IDs
- [ ] Implement node add command: create NodeAddCommand with Run(), use schema.WithArguments() for node-id and output positional args, schema.WithFields() for --type and --status flags
- [ ] Implement node edit command: create NodeEditCommand with Run(), use schema.WithArguments() with TypeStringList for node-ids (supports batch: node edit id1 id2 --status complete), schema.WithFields() for --status flag, handle status updates and completion timestamps for all nodes
- [ ] Implement node delete command: create NodeDeleteCommand with Run(), use schema.WithArguments() with TypeStringList for node-ids (supports batch: node delete id1 id2 --force), schema.WithFields() for --force flag, check blocked nodes before deletion for all nodes
- [ ] Implement graph command: create GraphCommand with RunIntoGlazeProcessor, use schema.NewGlazedSchema() for output formatting, schema.WithArguments() for optional goal-id, schema.WithFields() for --mermaid flag
- [ ] Implement goals command: create GoalsCommand with RunIntoGlazeProcessor, use schema.NewGlazedSchema() for output formatting, schema.WithFields() for --mermaid flag, output pending nodes as structured rows
- [ ] Implement history command: create HistoryCommand with RunIntoGlazeProcessor, use schema.NewGlazedSchema() for output formatting, schema.WithFields() for --limit, --since, --summary flags, parse relative time strings
- [ ] Implement search command: create SearchCommand with RunIntoGlazeProcessor, use schema.NewGlazedSchema() for output formatting, schema.WithArguments() for optional query, schema.WithFields() for all filter flags, implement tactic ranking logic
- [ ] Implement apply command: create ApplyCommand with Run(), use schema.WithArguments() for tactic-id, schema.WithFields() for --yes and --force flags, implement dependency checking and node creation logic
- [ ] Create root Cobra command: set up cmd/tactician/main.go with root command, create root.go files in each command group directory (pkg/commands/{init,node,graph,goals,history,search,apply}/root.go) to register command groups explicitly, register all groups in main.go using cli.BuildCobraCommandFromCommand(), configure CobraParserConfig with AppName for env var parsing
- [ ] Implement helper functions: create computeNodeStatus(), getPendingNodes(), parseRelativeTime(), rankTactics(), checkDependencies() matching JavaScript logic
- [ ] Add unit tests: create tests for each command, test database operations, test value decoding, test error handling
- [ ] Add integration tests: test full workflows (init → add node → search → apply → graph), test Mermaid output, test all flag combinations
- [ ] Implement LLM reranking (optional): create LLM section schema, implement reranker using OpenAI API matching JavaScript implementation, integrate with search command
