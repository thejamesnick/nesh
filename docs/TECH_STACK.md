# Current Technology Stack (As of 2026-04-09)

## Overview
Nesh is currently in the planning/design phase with no implementation yet. This document captures what we currently have and what we need to build.

## Current Assets
- **Documentation**: 
  - `nesh_script_product_technical_specification.md` - Complete product vision, philosophy, and technical spec
  - `About.txt` - Brief description
  - `nesh_technology_stack_plan.md` - Empty placeholder file

## Current Technology Stack (Planned)
Based on the technical specification:

### Core Language Runtime
- **Primary Implementation Language**: Go
  - Fast compilation
  - Strong concurrency support  
  - Cross-platform binaries
  - Suitable for CLI and system tools

### Parser
- Custom parser written in Go
- Tokenizer + grammar-based parsing
- Support for expressions, commands, and control flow

### CLI Interface
- Terminal-based REPL (planned)
- Script execution from files (planned)
- Interactive shell mode (planned)

### System Integration
- OS process execution (exec calls) (planned)
- Environment variable access (planned)
- File system operations (planned)

### Optional AI Integration Layer
- Structured command representation (JSON-like AST) (planned)
- API hooks for AI agents (planned)
- Natural language to Nesh translation layer (future) (planned)

## Components to Build (Per Spec)
1. **Lexer** - Converts raw input into tokens
2. **Parser** - Builds structured representation (AST or IR)
3. **Runtime / Executor** - Executes parsed instructions
4. **Standard Library** - Built-in commands: print, let, if, fn
5. **CLI Tool** - `nesh` command for running scripts and interactive mode
6. **Package / Module System** (Future) - Reusable scripts, import/export

## File Format
- **Extension**: `.nsh` (distinguishes from traditional `.sh` shell scripts)

## Current Implementation Status
- ❌ No code written yet
- ❌ No lexer implemented
- ❌ No parser implemented  
- ❌ No runtime/executor implemented
- ❌ No standard library built
- ❌ No CLI tool created
- ❌ No benchmarking infrastructure

## Immediate Next Steps
1. Set up Go project structure
2. Implement lexer
3. Implement parser
4. Build basic runtime with core built-ins
5. Create CLI wrapper
6. Establish benchmarking suite