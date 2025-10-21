# Requirements Clarity Command

## When to Use

Use `/clarif` when you have a vague requirement that needs systematic clarification to become implementation-ready.

## Command Syntax

```
/clarif <your requirement description>
```

## What This Command Does

Transforms vague requirements into actionable PRDs through:

1. **Initial Assessment** (0-100 clarity score)
2. **Interactive Q&A** (2-3 questions per round)
3. **Iterative Refinement** (until score ≥ 90)
4. **PRD Generation** (structured requirements document)

## Output Files

Generated in `./.claude/specs/{feature-name}/`:

- `clarification-log.md` - Complete Q&A history
- `prd.md` - Final product requirements document

## Example

```
/clarif I want to implement a user login feature
```

Claude will:
- Assess clarity (initial score: ~35/100)
- Ask 2-3 focused questions about login method, scope, tech stack
- Update score based on your answers
- Continue Q&A rounds until ≥ 90/100
- Generate complete PRD with acceptance criteria and execution phases

## When NOT to Use

Skip `/clarif` if your requirement already includes:
- Clear inputs/outputs
- Specified technology stack
- Defined acceptance criteria
- Technical constraints
- Edge case handling

## Pro Tips

1. Start with any level of detail - the command adapts
2. Answer 2-3 questions at a time (builds context progressively)
3. Review generated PRD before implementation
4. Use PRD as blueprint for development
