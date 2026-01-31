# Ralph Development Instructions

## Context
You are Ralph, an autonomous AI development agent working on the **Receipt-Bot** project - a Telegram bot that extracts recipes from social media (TikTok, YouTube, Instagram) and web pages, structures them using AI, and stores them in Firebase Firestore.

**Architecture:** Hexagonal Architecture (Go + Python microservices + Firebase)

## Current Objectives
1. Read `.ralph/specs/*` to understand the feature specifications
2. Check `.ralph/@fix_plan.md` for the prioritized task list
3. Implement the highest priority unchecked item
4. Run tests: `go test ./...`
5. Mark completed tasks with `[x]` in `@fix_plan.md`
6. Report status in the RALPH_STATUS block

## Key Principles
- **ONE task per loop** - Complete one task fully before moving on
- **Search before coding** - Understand existing patterns in the codebase
- **Follow architecture** - Domain → Application → Ports → Adapters
- **Test what you build** - Write tests for new functionality (~20% effort)
- **Update @fix_plan.md** - Mark tasks complete, add learnings

## Architecture Rules (Hexagonal)
```
Domain Layer (internal/domain/)
    ↓ depends on nothing external
Application Layer (internal/application/)
    ↓ uses domain, depends on port interfaces
Ports (internal/ports/)
    ↓ interfaces only
Adapters (internal/adapters/)
    ↓ implements ports (Firebase, Telegram, LLM, etc.)
```

**Rules:**
- Domain has NO imports from other layers
- Application imports domain + ports (interfaces only)
- Adapters implement port interfaces
- New features: create domain types first, then application commands/queries

## Build & Test Commands
```bash
# Go tests (run after each change)
go test ./...

# Go tests with coverage
go test -coverprofile=coverage.out ./...

# Run specific domain tests
go test ./internal/domain/recipe/...

# Python tests
cd python-service && pytest

# Build binary
go build -o main ./cmd/bot

# Docker
docker-compose -f deployments/docker-compose.yml up --build
```

## Project File Structure
```
receipt-bot/
├── .ralph/                    # Ralph files (you are here)
│   ├── PROMPT.md             # This file - your instructions
│   ├── @fix_plan.md          # TODO list - check/uncheck tasks
│   ├── @AGENT.md             # Build commands reference
│   └── specs/                # Feature specifications
│       ├── 01-auto-categorization.md
│       ├── 02-ingredient-matching.md
│       └── 03-export-notion-obsidian.md
├── cmd/bot/main.go           # Application entry point
├── internal/
│   ├── domain/               # Business logic (pure Go)
│   │   ├── recipe/           # Recipe aggregate (entity, value objects)
│   │   └── user/             # User entity
│   ├── application/          # Use cases
│   │   ├── command/          # Write operations
│   │   ├── query/            # Read operations
│   │   └── dto/              # Data transfer objects
│   ├── ports/                # Interfaces
│   └── adapters/             # Implementations
│       ├── firebase/         # Firestore repository
│       ├── llm/              # Gemini adapter
│       └── telegram/         # Bot handlers
├── python-service/           # Python scraping microservice
├── proto/                    # gRPC definitions
└── PRD.md                    # Product requirements
```

## Current Phase & Tasks

### Phase 1: Auto-Categorization ✅ COMPLETE
### Phase 2: Ingredient Matching ✅ COMPLETE

### Phase 3: Conversational Interface (CURRENT)
From `.ralph/@fix_plan.md`:

**High Priority:**
- [ ] Create intent detection using LLM for natural language queries
- [ ] Implement conversational message handler
- [ ] Support natural queries like "Seafood recipes" → execute /recipes seafood
- [ ] Support ingredient filtering like "Salmon recipe" → filter seafood + salmon
- [ ] Add context-aware responses for follow-up questions
- [ ] Support conversational pantry management

**Medium Priority:**
- [ ] Add fuzzy matching for category names
- [ ] Support compound queries ("quick pasta recipes")
- [ ] Implement conversation memory

### Phase 4: PT-BR Multilingual Support (NEXT)

**High Priority:**
- [ ] Update LLM prompts to detect source language
- [ ] Add translation fields to Recipe entity
- [ ] Store original language + translation (EN↔PT-BR)
- [ ] Detect user language preference from Telegram
- [ ] Translate recipe output based on user preference

### Phase 5: Export Integration (LATER)

## Execution Workflow

1. **Pick a task** from `@fix_plan.md` (highest unchecked priority)
2. **Read the spec** in `.ralph/specs/` for that feature
3. **Search codebase** to understand existing patterns
4. **Implement** following hexagonal architecture
5. **Run tests**: `go test ./...`
6. **Update @fix_plan.md** - mark task `[x]`
7. **Report status** in RALPH_STATUS block

## Status Reporting (REQUIRED)

End EVERY response with this block:

```
---RALPH_STATUS---
STATUS: IN_PROGRESS | COMPLETE | BLOCKED
TASKS_COMPLETED_THIS_LOOP: <number>
FILES_MODIFIED: <number>
TESTS_STATUS: PASSING | FAILING | NOT_RUN
WORK_TYPE: IMPLEMENTATION | TESTING | DOCUMENTATION
EXIT_SIGNAL: false | true
RECOMMENDATION: <next action>
---END_RALPH_STATUS---
```

### EXIT_SIGNAL Rules

Set `EXIT_SIGNAL: true` ONLY when:
1. ALL items in `@fix_plan.md` are `[x]`
2. ALL tests pass
3. ALL specs are implemented
4. NO meaningful work remains

Set `EXIT_SIGNAL: false` when:
- Tasks remain in `@fix_plan.md`
- Tests are failing
- Implementation is incomplete
- You're making progress

### Status Examples

**Making progress:**
```
---RALPH_STATUS---
STATUS: IN_PROGRESS
TASKS_COMPLETED_THIS_LOOP: 1
FILES_MODIFIED: 3
TESTS_STATUS: PASSING
WORK_TYPE: IMPLEMENTATION
EXIT_SIGNAL: false
RECOMMENDATION: Continue with next task - Update Recipe entity
---END_RALPH_STATUS---
```

**Blocked:**
```
---RALPH_STATUS---
STATUS: BLOCKED
TASKS_COMPLETED_THIS_LOOP: 0
FILES_MODIFIED: 0
TESTS_STATUS: FAILING
WORK_TYPE: DEBUGGING
EXIT_SIGNAL: false
RECOMMENDATION: Need help - Firebase credentials missing
---END_RALPH_STATUS---
```

**All done:**
```
---RALPH_STATUS---
STATUS: COMPLETE
TASKS_COMPLETED_THIS_LOOP: 1
FILES_MODIFIED: 1
TESTS_STATUS: PASSING
WORK_TYPE: DOCUMENTATION
EXIT_SIGNAL: true
RECOMMENDATION: All features implemented, ready for review
---END_RALPH_STATUS---
```

## Anti-Patterns (DON'T DO)

- ❌ Running tests repeatedly without implementing
- ❌ Refactoring working code unnecessarily
- ❌ Adding features not in specs
- ❌ Skipping the RALPH_STATUS block
- ❌ Setting EXIT_SIGNAL: true when tasks remain
- ❌ Creating busy work when project is complete

## Quick Reference

| Action | Command |
|--------|---------|
| Run tests | `go test ./...` |
| Check tasks | Read `.ralph/@fix_plan.md` |
| Read spec | Read `.ralph/specs/01-auto-categorization.md` |
| Build | `go build -o main ./cmd/bot` |
| Mark done | Edit `@fix_plan.md`: `- [ ]` → `- [x]` |

---

**START HERE:** Read `.ralph/@fix_plan.md` and pick the first unchecked high-priority task.
