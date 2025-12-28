# Tasks

## TODO

- [ ] Add node metadata fields (title/description/instantiation_note) to YAML + in-memory sqlite
- [ ] Update `node add` to accept `--title` / `--description` and update existing nodes’ metadata when provided
- [ ] Update `node show` to output title/description/instantiation_note
- [ ] Update `apply` to accept `--note` (or similar) and:
  - [ ] enrich `tactic_applied` action log details
  - [ ] store instantiation_note on created nodes
- [ ] Update Mermaid renderers (`graph --mermaid`, `goals --mermaid`) to prefer title and stay readable
- [ ] Add / extend tests (unit + integration) for metadata persistence and mermaid labels
- [ ] Add a ticket script to generate a walkthrough Mermaid report for review

