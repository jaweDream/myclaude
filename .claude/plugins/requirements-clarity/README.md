# Requirements Clarity Plugin

## Overview

This Claude Code plugin automatically detects vague requirements and transforms them into crystal-clear Product Requirements Documents (PRDs) through systematic clarification.

**Plugin vs Skill vs Command**:
- **Plugin**: Distributable package with metadata (`claude.json`)
- **Skill**: Auto-activated behavior (part of plugin)
- **Command**: Manual invocation (deprecated in favor of plugin)

## Installation

```bash
/plugin install requirements-clarity
```

Or add to your `.clauderc`:

```json
{
  "plugins": {
    "requirements-clarity": {
      "enabled": true
    }
  }
}
```

## How It Works

### Automatic Activation

The plugin activates when Claude detects:

1. **Vague Feature Requests**
   ```
   User: "add login feature"
   User: "implement payment system"
   User: "create user dashboard"
   ```

2. **Missing Technical Details**
   - No technology stack mentioned
   - No architecture or constraints specified
   - No integration points identified

3. **Incomplete Specifications**
   - No acceptance criteria
   - No success metrics
   - No edge cases or error handling

4. **Ambiguous Scope**
   - Unclear boundaries ("user management" - what exactly?)
   - No distinction between MVP and future features

### Clarification Process

```
User: "I want to implement a user login feature"
  ↓
Claude detects vague requirement
  ↓
Auto-activates requirements-clarity skill
  ↓
Initial assessment: 35/100 clarity score
  ↓
Round 1: Ask 2-3 targeted questions
  ↓
User responds
  ↓
Score update: 35 → 72
  ↓
Round 2: Continue clarifying gaps
  ↓
User responds
  ↓
Score update: 72 → 93 ✓ (≥90 threshold)
  ↓
Generate PRD files:
  - ./.claude/specs/user-login/prd.md
  - ./.claude/specs/user-login/clarification-log.md
```

## Scoring System (100 points)

| Dimension | Points | Criteria |
|-----------|--------|----------|
| **Functional Clarity** | 30 | Clear inputs/outputs (10), User interaction (10), Success criteria (10) |
| **Technical Specificity** | 25 | Tech stack (8), Integration points (8), Constraints (9) |
| **Implementation Completeness** | 25 | Edge cases (8), Error handling (9), Data validation (8) |
| **Business Context** | 20 | Problem statement (7), Target users (7), Success metrics (6) |

**Threshold**: ≥ 90 points required before PRD generation

## Output Structure

### 1. Clarification Log
`./.claude/specs/{feature-name}/clarification-log.md`

Documents the entire clarification conversation:
- Original requirement
- Each round of questions and answers
- Score progression
- Final assessment breakdown

### 2. Product Requirements Document
`./.claude/specs/{feature-name}/prd.md`

Structured PRD with four main sections:

#### Requirements Description
- Background: Business problem, target users, value proposition
- Feature Overview: Core functionality, boundaries, user scenarios
- Detailed Requirements: Inputs/outputs, interactions, data, edge cases

#### Design Decisions
- Technical Approach: Architecture, components, data storage, APIs
- Constraints: Performance, compatibility, security, scalability
- Risk Assessment: Technical, dependency, timeline risks

#### Acceptance Criteria
- Functional: Checklistable feature requirements
- Quality Standards: Code quality, testing, performance, security
- User Acceptance: UX, documentation, training

#### Execution Phases
- Phase 1: Preparation - Environment setup
- Phase 2: Core Development - Core implementation
- Phase 3: Integration & Testing - QA
- Phase 4: Deployment - Release

## Testing Guide

### Test Case 1: Vague Login Feature

**Input**:
```
"I want to implement a user login feature"
```

**Expected Behavior**:
1. Claude detects vague requirement
2. Announces activation of requirements-clarity skill
3. Shows initial score (~30-40/100)
4. Asks 2-3 questions about:
   - Login method (username+password, phone+OTP, OAuth?)
   - Functional scope (remember me, forgot password?)
   - Technology stack (backend language, database, auth method?)

**Expected Output**:
- Score improves to ~70+ after round 1
- Additional questions about security, error handling, performance
- Final score ≥ 90
- PRD generated in `./.claude/specs/user-login/`

### Test Case 2: Ambiguous E-commerce Feature

**Input**:
```
"add shopping cart to the website"
```

**Expected Behavior**:
1. Auto-activation (no tech stack, no UX details, no constraints)
2. Questions about:
   - Cart behavior (guest checkout? save for later? quantity limits?)
   - User experience (inline cart vs dedicated page?)
   - Backend integration (existing inventory system? payment gateway?)
   - Data persistence (session storage, database, local storage?)

**Expected Output**:
- Iterative clarification (2-3 rounds)
- Score progression: ~25 → ~65 → ~92
- PRD with concrete shopping cart specifications

### Test Case 3: Technical Implementation Request

**Input**:
```
"Refactor the authentication service to use JWT tokens"
```

**Expected Behavior**:
1. May NOT activate (already fairly specific)
2. If activates, asks about:
   - Token expiration strategy
   - Refresh token implementation
   - Migration plan from existing auth
   - Backward compatibility requirements

### Test Case 4: Clear Requirement (Should NOT Activate)

**Input**:
```
"Fix the null pointer exception in auth.go line 45 by adding a nil check before accessing user.Email"
```

**Expected Behavior**:
- Skill does NOT activate (requirement is already clear)
- Claude proceeds directly to implementation

## Benefits

1. **Proactive Quality Gate**: Prevents unclear specs from proceeding to implementation
2. **Zero Friction**: Users describe features naturally, no command syntax needed
3. **Context Awareness**: Claude recognizes ambiguity patterns automatically
4. **Persistent Mode**: Stays active throughout clarification conversation
5. **Better UX**: Natural conversation flow vs explicit command invocation

## Configuration

No configuration needed - the plugin is automatically discovered by Claude Code when installed.

**Plugin Metadata** (in claude.json):
```json
{
  "name": "requirements-clarity",
  "version": "1.0.0",
  "description": "Automatically clarify vague requirements into actionable PRDs",
  "components": {
    "skills": ["requirements-clarity"]
  }
}
```

## Troubleshooting

### Plugin Not Activating

**Problem**: Claude doesn't enter clarification mode for vague requirements

**Solutions**:
1. Verify plugin is installed: `/plugin list`
2. Check plugin is enabled in `.clauderc`
3. Ensure `instructions.md` exists in plugin directory
4. Try more explicit vague requirement: "add user feature"

### Premature PRD Generation

**Problem**: PRD generated before score reaches 90

**Solution**: This is a bug - instructions.md explicitly requires ≥90 threshold. Review the clarification log to see actual score.

### Over-Clarification

**Problem**: Claude asks too many questions for simple features

**Expected**: This is by design - better to over-clarify than under-specify. If the requirement is truly simple, answer questions quickly to reach 90+ score faster.

## Examples

### Example 1: Login Feature (Full Flow)

See full example in instructions.md under "Example Clarification Flow"

**Summary**:
- Input: "I want to implement a user login feature"
- Round 1: Login method, scope, tech stack → Score 35→72
- Round 2: Security, error handling, performance → Score 72→93
- Output: Complete PRD with bcrypt, JWT, PostgreSQL, Go backend

### Example 2: API Endpoint

**Input**: "create an API to get user profile"

**Round 1** (Score: 28/100):
```
Q1: What information should the API return? (name, email, avatar, preferences?)
Q2: Authentication required? (JWT, session, API key?)
Q3: Response format? (JSON, XML?) Any pagination?
```

**Round 2** (Score: 75/100):
```
Q1: Error handling: What if user not found? (404, custom error?)
Q2: Performance: Caching strategy? Expected QPS?
Q3: Privacy: Any fields that should be filtered based on requester?
```

**Round 3** (Score: 91/100):
```
PRD Generated:
- Endpoint: GET /api/v1/users/:id
- Auth: JWT required
- Response: JSON with name, email, avatar, bio
- Caching: Redis, 5min TTL
- Rate limit: 100 req/min per user
```

## References

- **Claude Code Plugins Documentation**: https://docs.claude.com/en/docs/claude-code/plugins
- **Original `/clarif` Command**: `development-essentials/commands/clarif.md`
- **Original Clarification Agent**: `development-essentials/agents/clarif-agent.md`

## Changelog

### v1.0.0 (2025-10-21)
- Converted skill to plugin format
- Added `claude.json` plugin metadata
- Translated all prompts to English (was mixed Chinese/English)
- Updated documentation for plugin distribution
- Maintained 100-point scoring system and PRD structure
- Compatible with Claude Code plugin system

---

**License**: MIT  
**Author**: stellarlink  
**Homepage**: https://github.com/cexll/myclaude
