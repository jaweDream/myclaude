---
name: requirements-clarity
description: Automatically detect vague requirements and transform them into crystal-clear Product Requirements Documents (PRDs) through systematic clarification
activation_triggers:
  - User describes a feature without technical details
  - Request lacks acceptance criteria or success metrics
  - Scope is ambiguous (e.g., "add authentication", "implement user management")
  - Missing technology stack or implementation constraints
  - No edge cases or error handling mentioned
  - Vague action verbs without specifics ("add", "create", "improve", "fix")
tools: Read, Write, Glob, Grep, TodoWrite
---

# Requirements Clarity Skill

## When to Activate

This skill should automatically activate when Claude detects:

1. **Vague Feature Requests**
   - User says: "add login feature", "implement payment", "create dashboard"
   - Missing: How, with what technology, what constraints?

2. **Missing Technical Context**
   - No technology stack mentioned
   - No integration points identified
   - No performance/security constraints

3. **Incomplete Specifications**
   - No acceptance criteria
   - No success metrics
   - No edge cases considered
   - No error handling mentioned

4. **Ambiguous Scope**
   - Unclear boundaries ("user management" - what exactly?)
   - No distinction between MVP and future enhancements
   - Missing "what's NOT included"

## Skill Behavior

Upon activation, this skill transforms vague requirements into actionable PRDs through systematic clarification.

### Core Principles

1. **Systematic Questioning**
   - Ask focused, specific questions
   - One category at a time (2-3 questions per round)
   - Build on previous answers
   - Avoid overwhelming users

2. **Quality-Driven Iteration**
   - Continuously assess clarity score (0-100)
   - Identify gaps systematically
   - Iterate until ≥ 90 points
   - Document all clarification rounds

3. **Actionable Output**
   - Generate concrete specifications
   - Include measurable acceptance criteria
   - Provide executable phases
   - Enable direct implementation

---

## Clarification Process

### Step 1: Initial Requirement Analysis

**Input**: User's requirement description

**Tasks**:
1. Parse and understand core requirement
2. Generate feature name (kebab-case format)
3. Create output directory: `./.claude/specs/{feature_name}/`
4. Perform initial clarity assessment (0-100)

**Assessment Rubric**:
```
功能清晰度 (Functional Clarity): /30 points
- Clear inputs/outputs: 10 pts
- User interaction defined: 10 pts
- Success criteria stated: 10 pts

技术具体性 (Technical Specificity): /25 points
- Technology stack mentioned: 8 pts
- Integration points identified: 8 pts
- Constraints specified: 9 pts

实现完整性 (Implementation Completeness): /25 points
- Edge cases considered: 8 pts
- Error handling mentioned: 9 pts
- Data validation specified: 8 pts

业务背景 (Business Context): /20 points
- Problem statement clear: 7 pts
- Target users identified: 7 pts
- Success metrics defined: 6 pts
```

**Initial Response Format**:
```markdown
我已经理解您的需求。让我帮您完善这个需求规格。

**当前清晰度评分**: X/100

**已明确的部分**:
- [List what's clear]

**需要澄清的部分**:
- [List gaps]

让我开始系统性地澄清这些内容...
```

### Step 2: Gap Analysis

Identify missing information across four dimensions:

**1. 功能范围 (Functional Scope)**
- What is the core functionality?
- What are the boundaries?
- What is out of scope?
- What are edge cases?

**2. 用户交互 (User Interaction)**
- How do users interact?
- What are the inputs?
- What are the outputs?
- What are success/failure scenarios?

**3. 技术约束 (Technical Constraints)**
- Performance requirements?
- Compatibility requirements?
- Security considerations?
- Scalability needs?

**4. 业务价值 (Business Value)**
- What problem does this solve?
- Who are the target users?
- What are success metrics?
- What is the priority?

### Step 3: Interactive Clarification

**Question Strategy**:
1. Start with highest-impact gaps
2. Ask 2-3 questions per round
3. Build context progressively
4. Use user's language
5. Provide examples when helpful

**Question Format**:
```markdown
我需要澄清以下几点以完善需求文档:

1. **[Category]**: [Specific question]?
   - 例如: [Example if helpful]

2. **[Category]**: [Specific question]?

3. **[Category]**: [Specific question]?

请提供您的答案,我会基于此继续完善 PRD。
```

**After Each User Response**:
1. Update clarity score
2. Document new information in clarification log
3. Identify remaining gaps
4. If score < 90: Continue with next round of questions
5. If score ≥ 90: Proceed to PRD generation

**Score Update Format**:
```markdown
感谢您的补充! 

**清晰度评分更新**: X/100 → Y/100

**新增明确的内容**:
- [Summarize new information]

**剩余需要澄清的点**:
- [List remaining gaps if score < 90]

[If score < 90: Continue with next round of questions]
[If score ≥ 90: "完美! 我现在将生成完整的 PRD 文档..."]
```

### Step 4: PRD Generation

Once clarity score ≥ 90, generate comprehensive PRD.

**Output Files**:

1. **Clarification Log**: `./.claude/specs/{feature_name}/clarification-log.md`
2. **Final PRD**: `./.claude/specs/{feature_name}/prd.md`

---

## PRD Document Structure

```markdown
# {Feature Name} - 产品需求文档 (PRD)

## 需求描述 (Requirements Description)

### 背景 (Background)
- **业务问题**: [描述要解决的业务问题]
- **目标用户**: [目标用户群体]
- **价值主张**: [此功能带来的价值]

### 功能概述 (Feature Overview)
- **核心功能**: [主要功能点列表]
- **功能边界**: [明确包含和不包含的内容]
- **用户场景**: [典型使用场景描述]

### 详细需求 (Detailed Requirements)
- **输入/输出**: [具体的输入输出规格]
- **用户交互**: [用户操作流程]
- **数据要求**: [数据结构和验证规则]
- **边界条件**: [边界情况处理]

## 设计决策 (Design Decisions)

### 技术方案 (Technical Approach)
- **架构选择**: [技术架构决策及理由]
- **关键组件**: [主要技术组件列表]
- **数据存储**: [数据模型和存储方案]
- **接口设计**: [API/接口规格]

### 约束条件 (Constraints)
- **性能要求**: [响应时间、吞吐量等]
- **兼容性**: [系统兼容性要求]
- **安全性**: [安全相关考虑]
- **可扩展性**: [未来扩展考虑]

### 风险评估 (Risk Assessment)
- **技术风险**: [潜在技术风险及缓解方案]
- **依赖风险**: [外部依赖及备选方案]
- **时间风险**: [进度风险及应对策略]

## 验收标准 (Acceptance Criteria)

### 功能验收 (Functional Acceptance)
- [ ] 功能点1: [具体验收条件]
- [ ] 功能点2: [具体验收条件]
- [ ] 功能点3: [具体验收条件]

### 质量标准 (Quality Standards)
- [ ] 代码质量: [代码规范和审查要求]
- [ ] 测试覆盖: [测试要求和覆盖率]
- [ ] 性能指标: [性能测试通过标准]
- [ ] 安全检查: [安全审查要求]

### 用户验收 (User Acceptance)
- [ ] 用户体验: [UX验收标准]
- [ ] 文档完整: [文档交付要求]
- [ ] 培训材料: [如需要,培训材料要求]

## 执行 Phase (Execution Phases)

### Phase 1: 准备阶段 (Preparation)
**目标**: 环境准备和技术验证
- [ ] 任务1: [具体任务描述]
- [ ] 任务2: [具体任务描述]
- **产出**: [阶段交付物]
- **时间**: [预估时间]

### Phase 2: 核心开发 (Core Development)
**目标**: 实现核心功能
- [ ] 任务1: [具体任务描述]
- [ ] 任务2: [具体任务描述]
- **产出**: [阶段交付物]
- **时间**: [预估时间]

### Phase 3: 集成测试 (Integration & Testing)
**目标**: 集成和质量保证
- [ ] 任务1: [具体任务描述]
- [ ] 任务2: [具体任务描述]
- **产出**: [阶段交付物]
- **时间**: [预估时间]

### Phase 4: 部署上线 (Deployment)
**目标**: 发布和监控
- [ ] 任务1: [具体任务描述]
- [ ] 任务2: [具体任务描述]
- **产出**: [阶段交付物]
- **时间**: [预估时间]

---

**文档版本**: 1.0  
**创建时间**: {timestamp}  
**澄清轮数**: {clarification_rounds}  
**质量评分**: {quality_score}/100
```

---

## Clarification Log Structure

`./.claude/specs/{feature_name}/clarification-log.md`

```markdown
# Requirements Clarification Log

## Original Requirement
[User's initial input]

## Clarification Rounds

### Round 1 (Score: X/100)
**Questions**:
1. ...
2. ...

**User Responses**:
1. ...
2. ...

**Score Update**: X → Y points

**Gaps Addressed**:
- [List what became clear]

### Round 2 (Score: Y/100)
...

## Final Clarity Assessment
- 功能清晰度 (Functional Clarity): X/30
- 技术具体性 (Technical Specificity): X/25
- 实现完整性 (Implementation Completeness): X/25
- 业务背景 (Business Context): X/20
- **Total**: X/100 ✓

## Key Decisions Made
1. [Important decision 1]
2. [Important decision 2]
```

---

## Quality Assurance

### Before PRD Generation
- [ ] Clarity score ≥ 90 points
- [ ] All four dimensions addressed
- [ ] Functional requirements complete
- [ ] Technical constraints identified
- [ ] Acceptance criteria defined
- [ ] Execution phases concrete

### PRD Completeness Check
- [ ] All sections filled with substance
- [ ] Checkboxes for acceptance criteria
- [ ] Concrete tasks in each phase
- [ ] Time estimates provided
- [ ] Deliverables specified
- [ ] Quality score documented

---

## Behavioral Guidelines

### DO
- Ask specific, targeted questions
- Build on previous answers
- Provide examples to guide users
- Maintain conversational tone
- Document all clarification rounds
- Use Chinese for section headers
- Generate concrete specifications
- Stay in clarification mode until score ≥ 90

### DON'T
- Ask all questions at once
- Make assumptions without confirmation
- Generate PRD before 90+ score
- Skip any required sections
- Use vague or abstract language
- Proceed without user responses
- Exit skill mode prematurely

---

## Example Clarification Flow

```
User: "我要做一个用户登录功能"

Claude: (Detects vague requirement → Auto-activates requirements-clarity skill)

我已经理解您的需求。让我帮您完善这个登录功能的需求规格。

**当前清晰度评分**: 35/100

**已明确的部分**:
- 需要实现用户登录功能

**需要澄清的部分**:
- 登录方式和技术实现
- 功能边界和用户交互
- 安全性和性能要求

让我开始系统性地澄清这些内容:

---

我需要澄清以下几点以完善需求文档:

1. **用户交互**: 用户将通过什么方式登录?
   - 例如: 用户名+密码, 手机号+验证码, 第三方OAuth (Google/GitHub)等

2. **功能范围**: 除了基本登录,是否需要:
   - 记住登录状态 (Remember Me)?
   - 忘记密码功能?
   - 多设备登录管理?

3. **技术约束**: 
   - 后端技术栈是什么? (Go/Node/Python/Java等)
   - 是否已有用户数据库?
   - 是否需要JWT token认证?

请提供您的答案,我会基于此继续完善 PRD。

---

[User responds: "用户名+密码, 需要记住登录和忘记密码, 后端用Go, 已有PostgreSQL, 需要JWT"]

---

Claude: 感谢您的补充! 

**清晰度评分更新**: 35/100 → 72/100

**新增明确的内容**:
- 登录方式: 用户名+密码
- 功能范围: 基本登录 + 记住登录 + 忘记密码
- 技术栈: Go后端 + PostgreSQL + JWT认证

**剩余需要澄清的点**:
- 安全性要求
- 错误处理策略
- 性能指标

继续澄清:

1. **安全性**: 
   - 密码加密方式有要求吗? (bcrypt/argon2/scrypt)
   - 是否需要限制登录尝试次数? (防暴力破解)
   - JWT token过期时间? (例如: 24小时, 7天)
   - Remember Me token有效期?

2. **错误处理**:
   - 登录失败时如何提示用户? (显示具体错误 vs 模糊提示)
   - 是否需要审计日志记录登录行为?

3. **性能要求**:
   - 预期并发登录用户数?
   - 响应时间要求? (例如: 登录接口 < 500ms)

请继续补充,我们快完成了!

---

[User responds with security and performance details]

---

Claude: 完美! 

**清晰度评分更新**: 72/100 → 93/100 ✓

**新增明确的内容**:
- 安全: bcrypt加密, 5次失败锁定15分钟, JWT 24小时过期
- 日志: 记录所有登录尝试
- 性能: 支持1000并发, 响应 < 300ms

需求已经非常清晰,我现在将生成完整的 PRD 文档...

[Creates ./.claude/specs/user-login/ directory]
[Generates clarification-log.md]
[Generates prd.md]

✅ **PRD 已生成**: 

1. **需求文档**: `./.claude/specs/user-login/prd.md`
2. **澄清记录**: `./.claude/specs/user-login/clarification-log.md`

**文档概览**:
- 需求描述: 完整的功能、用户场景、技术方案
- 设计决策: Go + PostgreSQL + JWT + bcrypt架构
- 验收标准: 12项功能验收 + 8项质量标准
- 执行Phase: 4个阶段,预估2-3周完成

您可以查看文档并确认是否需要调整。如果满意,我可以立即开始实施! 🚀
```

---

## Success Criteria

- Clarity score ≥ 90/100
- All PRD sections complete with substance
- Acceptance criteria checklistable (使用 `- [ ]` 格式)
- Execution phases actionable with concrete tasks
- User approves final PRD
- Ready for development handoff

---

## Important Notes

1. **Proactive Activation**: Don't wait for `/clarif` command - activate automatically when you detect vague requirements
2. **Persistent Mode**: Stay in clarification mode throughout the conversation until PRD is complete
3. **Never Skip**: Always iterate until ≥ 90 score, never generate PRD prematurely
4. **Document Everything**: Save both clarification-log.md and prd.md
5. **User Language**: Use Chinese for section headers, mixed Chinese/English for content
6. **Concrete Specifications**: Avoid vague language, make everything measurable and actionable
7. **Build Context**: Each question round builds on previous answers
8. **Quality Gate**: Requirements clarity is a quality gate - prevent unclear specs from proceeding to implementation
