---
description: Interactive TODO management for daily tasks and fixes outside feature workflows
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Overview

This command manages a simple TODO list in `.cursor/TODO.md` for daily tasks that fall outside the formal feature development workflow. Use it to track bugs, tech debt, quick fixes, and other ad-hoc tasks.

**Key Features**:
- No dependencies on feature directories or prerequisite scripts
- Interactive mode for managing TODOs
- Simple format: `- [ ] TODO001 [Category] Description with file/path`
- Persistent storage in `.cursor/TODO.md`

## Execution Steps

### 1. Determine Mode

Check if `$ARGUMENTS` is empty or contains content:

- **If $ARGUMENTS is empty**: Enter **Interactive Mode**
- **If $ARGUMENTS has content**: Enter **Quick Add Mode**

### 2. Load Existing TODOs

Read `.cursor/TODO.md` if it exists. Parse to extract:
- All existing TODO items
- Highest TODO ID number (e.g., if TODO015 exists, next ID is TODO016)
- Current categories in use

**File Format**:
```markdown
# TODO List

## Active Tasks

- [ ] TODO001 [Bug] Fix tournament API type mismatch in tournamentApi.ts
- [ ] TODO002 [TechDebt] Refactor player confirmation logic in services/competition/

## Completed Tasks

- [x] TODO003 [Refactor] Clean up unused imports in host components
```

If file doesn't exist, start with empty list and ID sequence at TODO001.

### 3a. Interactive Mode (when $ARGUMENTS is empty)

Present the following options to the user:

```
═══════════════════════════════════════
         TODO Management Menu
═══════════════════════════════════════

Current Stats: X active, Y completed

Options:
  1. Add new TODO
  2. View all TODOs
  3. View by category
  4. Mark as complete
  5. Remove TODO
  6. Exit

What would you like to do?
```

Then execute the selected action:

#### Option 1: Add new TODO
1. Ask: "Select category or type custom:"
   - Show: [Bug], [Feature], [TechDebt], [Refactor], [Docs], [Config], [Test], [Custom]
2. If Custom, ask: "Enter category name:"
3. Ask: "Enter TODO description (include file paths if relevant):"
4. Generate next TODO ID
5. Add to `.cursor/TODO.md` in Active Tasks section
6. Confirm: "✓ Added TODO### [Category] Description"

#### Option 2: View all TODOs
Display all TODOs organized by status:
```
Active Tasks (X items):
- [ ] TODO001 [Bug] Description
- [ ] TODO002 [TechDebt] Description

Completed Tasks (Y items):
- [x] TODO003 [Refactor] Description
```

#### Option 3: View by category
1. List all categories found in the file
2. Ask: "Select category to filter:"
3. Display only TODOs matching that category

#### Option 4: Mark as complete
1. Display all active (unchecked) TODOs
2. Ask: "Enter TODO ID to mark complete (e.g., TODO001):"
3. Validate ID exists and is active
4. Update checkbox from `- [ ]` to `- [x]`
5. Move item to Completed Tasks section
6. Confirm: "✓ Marked TODO### as complete"

#### Option 5: Remove TODO
1. Display all TODOs
2. Ask: "Enter TODO ID to remove (e.g., TODO001):"
3. Validate ID exists
4. Confirm: "Are you sure you want to remove TODO###? (yes/no)"
5. If yes, delete the line from file
6. Confirm: "✓ Removed TODO###"

#### Option 6: Exit
End the command execution.

**Loop behavior**: After completing an action (except Exit), return to the menu.

### 3b. Quick Add Mode (when $ARGUMENTS has content)

Parse `$ARGUMENTS` to extract category and description:

**Pattern 1**: Category is explicitly provided in brackets
- Input: `[Bug] Fix tournament API type mismatch in tournamentApi.ts`
- Category: `Bug`
- Description: `Fix tournament API type mismatch in tournamentApi.ts`

**Pattern 2**: Category is NOT provided
- Input: `Fix tournament API type mismatch in tournamentApi.ts`
- Ask: "Select category for this TODO:"
  - Show: [Bug], [Feature], [TechDebt], [Refactor], [Docs], [Config], [Test], [Custom]
- If Custom, ask: "Enter category name:"
- Category: Selected/entered category
- Description: Original input

**Pattern 3**: Multiple words with first word potentially being category
- Input: `Bug: Fix tournament API`
- Try to detect if first word matches known category (case-insensitive)
- If match, extract category and use rest as description
- Otherwise, treat as Pattern 2

Then:
1. Generate next TODO ID
2. Format: `- [ ] TODO### [Category] Description`
3. Add to Active Tasks section in `.cursor/TODO.md`
4. Confirm: "✓ Added TODO### [Category] Description"

### 4. File Management

**Creating/Updating `.cursor/TODO.md`**:

- If file doesn't exist, create with template:
```markdown
# TODO List

## Active Tasks

## Completed Tasks
```

- When adding items:
  - New active items go under `## Active Tasks` section
  - Maintain sequential order by ID
  - Keep proper markdown spacing (blank line between sections)

- When marking complete:
  - Move item from Active Tasks to Completed Tasks
  - Change `- [ ]` to `- [x]`
  - Maintain chronological order in Completed section (most recent at bottom)

- When removing:
  - Delete the entire line
  - Do NOT renumber remaining items (IDs are immutable)

**ID Generation**:
1. Scan all existing TODOs (both active and completed)
2. Extract numeric part (e.g., TODO015 → 15)
3. Find maximum number
4. New ID = max + 1, formatted as TODO### (zero-padded to 3 digits)
5. Examples: TODO001, TODO015, TODO123

### 5. Category Guidelines

**Default Categories**:
- `[Bug]` - Something broken that needs fixing
- `[Feature]` - New functionality to implement
- `[TechDebt]` - Code quality improvements, cleanup
- `[Refactor]` - Code restructuring without changing behavior
- `[Docs]` - Documentation updates or additions
- `[Config]` - Configuration changes (build, deploy, env)
- `[Test]` - Test-related tasks (add tests, fix tests)

**Custom Categories**: Users can add any category they want. Accept and use as-is.

### 6. Output & Reporting

After each operation, provide clear feedback:

**Adding TODO**:
```
✓ Added TODO012 [Bug] Fix tournament API type mismatch
  Location: .cursor/TODO.md
  Total active TODOs: 5
```

**Marking Complete**:
```
✓ Marked TODO007 as complete
  Total active TODOs: 4
  Total completed TODOs: 3
```

**Removing TODO**:
```
✓ Removed TODO009 [TechDebt] Refactor old code
  Total active TODOs: 4
```

**Viewing**:
Display the requested TODOs in a clean, readable format with counts.

## Best Practices

1. **Include file paths**: When describing TODOs, include specific file paths or locations
   - Good: `Fix type mismatch in frontends/winspire-app/src/features/host/api/tournamentApi.ts`
   - Bad: `Fix type mismatch`

2. **Be specific**: Clear descriptions make it easier to address later
   - Good: `Add error handling for failed API calls in PlayerConfirmation component`
   - Bad: `Fix errors`

3. **Use appropriate categories**: This helps filter and prioritize work

4. **Don't reuse IDs**: Even after deletion, never reuse TODO IDs to maintain history

5. **Regular cleanup**: Periodically review and remove completed TODOs to keep the list manageable

## Example Usage

**Interactive Mode**:
```
User: /todo.fix
Assistant: [Presents menu, guides through selection]
```

**Quick Add with Category**:
```
User: /todo.fix [Bug] Tournament API returns wrong player count in TournamentDetailPage.tsx
Assistant: ✓ Added TODO008 [Bug] Tournament API returns wrong player count in TournamentDetailPage.tsx
```

**Quick Add without Category**:
```
User: /todo.fix Investigate why websocket disconnects randomly
Assistant: Select category for this TODO:
  [Bug], [Feature], [TechDebt], [Refactor], [Docs], [Config], [Test], [Custom]
User: Bug
Assistant: ✓ Added TODO009 [Bug] Investigate why websocket disconnects randomly
```

## Notes

- This is a lightweight tool for daily task tracking, NOT a replacement for formal project management
- For feature development, use `speckit.tasks.md` instead
- The TODO list is stored locally in your workspace and not version controlled (unless you choose to add it to git)
- IDs are permanent and never reused to maintain a clear history




