# Requirements Clarity Skill

## Overview

This Claude Skill automatically detects vague requirements and transforms them into crystal-clear Product Requirements Documents (PRDs) through systematic clarification.

**Key Difference from `/clarif` Command**:
- **Command**: User must type `/clarif <requirement>` explicitly
- **Skill**: Claude automatically detects unclear requirements and activates clarification mode

## How It Works

### Automatic Activation

The skill activates when Claude detects:

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
User: "我要做一个用户登录功能"
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
| **功能清晰度** (Functional Clarity) | 30 | Clear inputs/outputs (10), User interaction (10), Success criteria (10) |
| **技术具体性** (Technical Specificity) | 25 | Tech stack (8), Integration points (8), Constraints (9) |
| **实现完整性** (Implementation Completeness) | 25 | Edge cases (8), Error handling (9), Data validation (8) |
| **业务背景** (Business Context) | 20 | Problem statement (7), Target users (7), Success metrics (6) |

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

#### 需求描述 (Requirements Description)
- 背景 (Background): Business problem, target users, value proposition
- 功能概述 (Feature Overview): Core functionality, boundaries, user scenarios
- 详细需求 (Detailed Requirements): Inputs/outputs, interactions, data, edge cases

#### 设计决策 (Design Decisions)
- 技术方案 (Technical Approach): Architecture, components, data storage, APIs
- 约束条件 (Constraints): Performance, compatibility, security, scalability
- 风险评估 (Risk Assessment): Technical, dependency, timeline risks

#### 验收标准 (Acceptance Criteria)
- 功能验收 (Functional): Checklistable feature requirements
- 质量标准 (Quality): Code quality, testing, performance, security
- 用户验收 (User): UX, documentation, training

#### 执行 Phase (Execution Phases)
- Phase 1: 准备阶段 (Preparation) - Environment setup
- Phase 2: 核心开发 (Core Development) - Core implementation
- Phase 3: 集成测试 (Integration & Testing) - QA
- Phase 4: 部署上线 (Deployment) - Release

## Testing Guide

### Test Case 1: Vague Login Feature

**Input**:
```
"我要做一个用户登录功能"
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

## Comparison: Command vs Skill

| Aspect | `/clarif` Command | Requirements-Clarity Skill |
|--------|-------------------|----------------------------|
| **Activation** | Manual: `/clarif <requirement>` | Automatic: Claude detects vague specs |
| **User Awareness** | Must know command exists | Transparent, no learning curve |
| **Workflow** | User → Type command → Clarification | User → Express need → Auto-clarification |
| **Discoverability** | Requires documentation | Claude suggests when appropriate |
| **Use Case** | Explicit requirements review | Proactive quality gate |
| **File Location** | `commands/clarif.md` + `agents/clarif-agent.md` | `.claude/skills/requirements-clarity/SKILL.md` |

## Benefits of Skill Approach

1. **Proactive Quality Gate**: Prevents unclear specs from proceeding to implementation
2. **Zero Friction**: Users describe features naturally, no command syntax needed
3. **Context Awareness**: Claude recognizes ambiguity patterns automatically
4. **Persistent Mode**: Stays active throughout clarification conversation
5. **Better UX**: Natural conversation flow vs explicit command invocation

## Configuration

No configuration needed - the skill is automatically discovered by Claude Code when present in `.claude/skills/requirements-clarity/`.

**Skill Metadata** (in SKILL.md frontmatter):
```yaml
name: requirements-clarity
description: Automatically clarify vague requirements into actionable PRDs
activation_triggers:
  - User describes feature without technical details
  - Request lacks acceptance criteria
  - Scope is ambiguous
  - Missing technology stack
tools: Read, Write, Glob, Grep, TodoWrite
```

## Troubleshooting

### Skill Not Activating

**Problem**: Claude doesn't enter clarification mode for vague requirements

**Solutions**:
1. Verify `.claude/skills/requirements-clarity/SKILL.md` exists
2. Check YAML frontmatter is valid
3. Ensure activation_triggers are defined
4. Try more explicit vague requirement: "add user feature"

### Premature PRD Generation

**Problem**: PRD generated before score reaches 90

**Solution**: This is a bug - SKILL.md explicitly requires ≥90 threshold. Review the clarification log to see actual score.

### Over-Clarification

**Problem**: Claude asks too many questions for simple features

**Expected**: This is by design - better to over-clarify than under-specify. If the requirement is truly simple, answer questions quickly to reach 90+ score faster.

## Migration from `/clarif` Command

The `/clarif` command in `development-essentials/commands/clarif.md` can coexist with this skill:

- **Skill**: Automatic activation for new, unclear requirements
- **Command**: Explicit review of existing requirements

**Recommended Workflow**:
1. User describes feature naturally
2. Skill auto-activates and generates PRD
3. (Optional) User runs `/clarif <existing-prd>` to review/refine

## Examples

### Example 1: Login Feature (Full Flow)

See full example in SKILL.md under "Example Clarification Flow"

**Summary**:
- Input: "我要做一个用户登录功能"
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

- **Claude Skills Documentation**: https://docs.claude.com/en/docs/claude-code/skills
- **Anthropic Skills Announcement**: https://www.anthropic.com/news/skills
- **Original `/clarif` Command**: `development-essentials/commands/clarif.md`
- **Original Clarification Agent**: `development-essentials/agents/clarif-agent.md`

## Changelog

### v1.0 (2025-10-20)
- Initial skill implementation
- Ported clarification logic from `/clarif` command
- Added automatic activation triggers
- Implemented 100-point scoring system
- Created bilingual PRD template (需求描述/设计决策/验收标准/执行Phase)
- Added comprehensive test cases and examples
