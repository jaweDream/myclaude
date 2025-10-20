# Advanced AI Agents Guide

> GPT-5 deep reasoning integration for complex analysis and architectural decisions

## 🎯 Overview

The Advanced AI Agents plugin provides access to GPT-5's deep reasoning capabilities through the `gpt5` agent, designed for complex problem-solving that requires multi-step thinking and comprehensive analysis.

## 🤖 GPT-5 Agent

### Capabilities

The `gpt5` agent excels at:

- **Architectural Analysis**: Evaluating system designs and scalability concerns
- **Strategic Planning**: Breaking down complex initiatives into actionable plans
- **Trade-off Analysis**: Comparing multiple approaches with detailed pros/cons
- **Problem Decomposition**: Breaking complex problems into manageable components
- **Deep Reasoning**: Multi-step logical analysis for non-obvious solutions
- **Technology Evaluation**: Assessing technologies, frameworks, and tools

### When to Use

**Use GPT-5 agent** when:
- Problem requires deep, multi-step reasoning
- Multiple solution approaches need evaluation
- Architectural decisions have long-term impact
- Trade-offs are complex and multifaceted
- Standard agents provide insufficient depth

**Use standard agents** when:
- Task is straightforward implementation
- Requirements are clear and well-defined
- Quick turnaround is priority
- Problem is domain-specific (code, tests, etc.)

## 🚀 Usage

### Via `/think` Command

The easiest way to access GPT-5:

```bash
/think "Analyze scalability bottlenecks in current microservices architecture"
/think "Evaluate migration strategy from monolith to microservices"
/think "Design data synchronization approach for offline-first mobile app"
```

### Direct Agent Invocation

For advanced usage:

```bash
# Use @gpt5 to invoke the agent directly
@gpt5 "Complex architectural question or analysis request"
```

## 💡 Example Use Cases

### 1. Architecture Evaluation

```bash
/think "Current system uses REST API with polling for real-time updates. 
Evaluate whether to migrate to WebSocket, Server-Sent Events, or GraphQL 
subscriptions. Consider: team experience, existing infrastructure, client 
support, scalability, and implementation effort."
```

**GPT-5 provides**:
- Detailed analysis of each option
- Pros and cons for your specific context
- Migration complexity assessment
- Performance implications
- Recommended approach with justification

### 2. Migration Strategy

```bash
/think "Plan migration from PostgreSQL to multi-region distributed database. 
System has 50M users, 200M rows, 1000 req/sec. Must maintain 99.9% uptime. 
What's the safest migration path?"
```

**GPT-5 provides**:
- Step-by-step migration plan
- Risk assessment for each phase
- Rollback strategies
- Data consistency approaches
- Timeline estimation

### 3. Problem Decomposition

```bash
/think "Design a recommendation engine that learns user preferences, handles 
cold start, provides explainable results, and scales to 10M users. Break this 
down into implementation phases with clear milestones."
```

**GPT-5 provides**:
- Problem breakdown into components
- Phased implementation plan
- Technical approach for each phase
- Dependencies between phases
- Success criteria and metrics

### 4. Technology Selection

```bash
/think "Choosing between Redis, Memcached, and Hazelcast for distributed 
caching. System needs: persistence, pub/sub, clustering, and complex data 
structures. Existing stack: Java, Kubernetes, AWS."
```

**GPT-5 provides**:
- Comparison matrix across requirements
- Integration considerations
- Operational complexity analysis
- Cost implications
- Recommendation with rationale

### 5. Performance Optimization

```bash
/think "API response time increased from 100ms to 800ms after scaling from 
100 to 10,000 users. Database queries look optimized. What are the likely 
bottlenecks and systematic approach to identify them?"
```

**GPT-5 provides**:
- Hypothesis generation (N+1 queries, connection pooling, etc.)
- Systematic debugging approach
- Profiling strategy
- Likely root causes ranked by probability
- Optimization recommendations

## 🎨 Integration with BMAD

### Enhanced Code Review

BMAD's `bmad-review` agent can optionally use GPT-5 for deeper analysis:

**Configuration**:
```bash
# Enable enhanced review mode (via environment or BMAD config)
BMAD_REVIEW_MODE=enhanced /bmad-pilot "feature description"
```

**What changes**:
- Standard review: Fast, focuses on code quality and obvious issues
- Enhanced review: Deep analysis including:
  - Architectural impact
  - Security implications
  - Performance considerations
  - Scalability concerns
  - Design pattern appropriateness

### Architecture Phase Support

Use `/think` during BMAD architecture phase:

```bash
# Start BMAD workflow
/bmad-pilot "E-commerce platform with real-time inventory"

# During Architecture phase, get deep analysis
/think "Evaluate architecture approaches for real-time inventory 
synchronization across warehouses, online store, and mobile apps"

# Continue with BMAD using insights
```

## 📋 Best Practices

### 1. Provide Complete Context

**❌ Insufficient**:
```bash
/think "Should we use microservices?"
```

**✅ Complete**:
```bash
/think "Current monolith: 100K LOC, 8 developers, 50K users, 200ms avg 
response time. Pain points: slow deployments (1hr), difficult to scale 
components independently. Should we migrate to microservices? What's the 
ROI and risk?"
```

### 2. Ask Specific Questions

**❌ Too broad**:
```bash
/think "How to build a scalable system?"
```

**✅ Specific**:
```bash
/think "Current system handles 1K req/sec. Need to scale to 10K. Bottleneck 
is database writes. Evaluate: sharding, read replicas, CQRS, or caching. 
Database: PostgreSQL, stack: Node.js, deployment: Kubernetes."
```

### 3. Include Constraints

Always mention:
- Team skills and size
- Timeline and budget
- Existing infrastructure
- Business requirements
- Technical constraints

**Example**:
```bash
/think "Design real-time chat system. Constraints: team of 3 backend 
developers (Node.js), 6-month timeline, AWS deployment, must integrate 
with existing REST API, budget for managed services OK."
```

### 4. Request Specific Outputs

Tell GPT-5 what format you need:

```bash
/think "Compare Kafka vs RabbitMQ for event streaming. 
Provide: comparison table, recommendation, migration complexity from current 
RabbitMQ setup, and estimated effort in developer-weeks."
```

### 5. Iterate and Refine

Follow up for deeper analysis:

```bash
# Initial question
/think "Evaluate caching strategies for user profile API"

# Follow-up based on response
/think "You recommended Redis with write-through caching. How to handle 
cache invalidation when user updates profile from mobile app?"
```

## 🔧 Technical Details

### Sequential Thinking

GPT-5 agent uses sequential thinking for complex problems:

1. **Problem Understanding**: Clarify requirements and constraints
2. **Hypothesis Generation**: Identify possible solutions
3. **Analysis**: Evaluate each option systematically
4. **Trade-off Assessment**: Compare pros/cons
5. **Recommendation**: Provide justified conclusion

### Reasoning Transparency

GPT-5 shows its thinking process:
- Assumptions made
- Factors considered
- Why certain options were eliminated
- Confidence level in recommendations

## 🎯 Comparison: GPT-5 vs Standard Agents

| Aspect | GPT-5 Agent | Standard Agents |
|--------|-------------|-----------------|
| **Depth** | Deep, multi-step reasoning | Focused, domain-specific |
| **Speed** | Slower (comprehensive analysis) | Faster (direct implementation) |
| **Use Case** | Strategic decisions, architecture | Implementation, coding, testing |
| **Output** | Analysis, recommendations, plans | Code, tests, documentation |
| **Best For** | Complex problems, trade-offs | Clear tasks, defined scope |
| **Invocation** | `/think` or `@gpt5` | `/code`, `/test`, etc. |

## 📚 Related Documentation

- **[BMAD Workflow](BMAD-WORKFLOW.md)** - Integration with full agile workflow
- **[Development Commands](DEVELOPMENT-COMMANDS.md)** - Standard command reference
- **[Quick Start Guide](QUICK-START.md)** - Get started quickly

## 💡 Advanced Patterns

### Pre-Implementation Analysis

```bash
# 1. Deep analysis with GPT-5
/think "Design approach for X with constraints Y and Z"

# 2. Use analysis in BMAD workflow
/bmad-pilot "Implement X based on approach from analysis"
```

### Architecture Validation

```bash
# 1. Get initial architecture from BMAD
/bmad-pilot "Feature X"  # Generates 02-system-architecture.md

# 2. Validate with GPT-5
/think "Review architecture in .claude/specs/feature-x/02-system-architecture.md
Evaluate for scalability, security, and maintainability"

# 3. Refine architecture based on feedback
```

### Decision Documentation

```bash
# Use GPT-5 to document architectural decisions
/think "Document decision to use Event Sourcing for order management.
Include: context, options considered, decision rationale, consequences,
and format as Architecture Decision Record (ADR)"
```

---

**Advanced AI Agents** - Deep reasoning for complex problems that require comprehensive analysis.
