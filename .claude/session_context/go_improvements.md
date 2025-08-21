# Go Parser Improvement Sprint Plan (Pragmatic Version)

## Updated Project Status - Keep It Simple

- **Current Completion**: 82% (Grade: B+) - Already working well!
- **Target**: 100% Production-Ready with Go idioms
- **Timeline**: 2 weeks (reduced from 3)
- **Philosophy**: Fix what's broken, improve what matters, skip the rest
- **Approach**: Incremental improvements, measure before optimizing

## Pragmatic Sprint Execution Strategy

### Phase 1: Fix Real Problems (Days 1-3)

**Agent 1 (debugger)**: Fix Actual Bugs

- [ ] Fix test failures in date formatting that actually break functionality
- [ ] Ensure all packages compile cleanly
- [ ] Fix only tests that represent real bugs (not just test issues)

**Agent 2 (elite-tdd-developer)**: Simple Error Improvements

- [ ] Use `fmt.Errorf` with `%w` for proper error wrapping (Go 1.13+ standard)
- [ ] Add 3-5 sentinel errors for common cases only:

  ```go
  var (
      ErrInvalidURL = errors.New("invalid URL")
      ErrResourceNotFound = errors.New("resource not found")
      ErrExtractionFailed = errors.New("extraction failed")
  )
  ```

- [ ] Skip custom error package - standard library is fine

**Agent 3 (elite-tdd-developer)**: Basic Context Support

- [ ] Add `ParseWithContext(ctx context.Context, url string)` method
- [ ] Keep existing `Parse()` method - it can call ParseWithContext with background context
- [ ] Add context only to long-running operations (HTTP calls, expensive extractions)
- [ ] Skip comprehensive propagation - add it where it matters

### Phase 2: Simple Improvements (Days 4-6)

**Agent 4 (elite-tdd-developer)**: Remove Duplication (Only Where Obvious)

- [ ] Extract shared HTTP configuration to one place (3 duplications is enough)
- [ ] Keep manual extractor registration - it's explicit and works
- [ ] Create 2-3 helper functions for truly repetitive code
- [ ] Skip complete overhauls - incremental improvement is fine

**Agent 5 (elite-tdd-developer)**: Complete Missing Features

- [ ] Implement missing cleaners (lead-image-url, resolve-split-title)
- [ ] Fix any actually broken extractors
- [ ] Skip "architectural overhauls" - current architecture works

**Agent 6 (elite-tdd-developer)**: Measured Performance Improvements

- [ ] First: Benchmark current performance
- [ ] Only optimize if benchmarks show real bottlenecks
- [ ] Simple improvements only:

  ```go
  // If regex compilation is actually slow:
  var (
      dateRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
      // ... other frequently used patterns
  )
  ```

- [ ] Skip parallel extraction unless benchmarks show it's needed
- [ ] Skip zero-allocation optimizations unless profiling shows allocation issues

### Phase 3: Polish and Ship (Days 7-10)

**Agents 1-2**: Practical Testing

- [ ] Fix failing tests that represent real bugs
- [ ] Add tests for new functionality only
- [ ] Target 70% coverage (not 85%) - focus on critical paths
- [ ] Skip comprehensive benchmark suite - add benchmarks only for suspected slow paths

**Agents 3-4**: Simple Documentation

- [ ] Document public API changes
- [ ] Add practical usage examples
- [ ] Skip migration guides and comprehensive docs - README updates are enough

## What We're NOT Doing (YAGNI)

### ❌ **Skip These Over-Engineered Features:**

1. **ParallelExtractor with worker pools** - Current extraction is fast enough
2. **Complete error system rewrite** - Standard library errors are fine
3. **Auto-registration with reflection** - Manual registration is explicit and debuggable
4. **Metrics, circuit breakers, rate limiting** - Let users add their own if needed
5. **Zero code duplication** - Some duplication is fine for clarity
6. **Complete architectural overhauls** - Current architecture works
7. **Comprehensive everything** - Good enough is good enough

## Realistic Success Metrics

### Week 1 (Days 1-6): Core Improvements

- ✅ Actual bugs fixed
- ✅ Context support where needed
- ✅ Simple error improvements
- ✅ Obvious duplication removed
- ✅ Missing features completed

### Week 2 (Days 7-10): Polish

- ✅ 70% test coverage on critical paths
- ✅ Performance measured and improved where needed
- ✅ Documentation updated
- ✅ Ready to ship

## Simple Coordination

- **Daily**: 5-minute check-in on Slack (not 30 minutes of meetings)
- **Branches**: Feature branches, merge when ready
- **Integration**: Test locally, merge if it works

## Actual Time Savings

**Original over-engineered plan**: 3 weeks (15 days)
**Pragmatic plan**: 2 weeks (10 days)
**Time saved**: 1 week (33% reduction)

**Effort comparison**:

- Original: 80 agent-days of complex work
- Pragmatic: 40 agent-days of focused work
- Effort saved: 50%

## The 80/20 Rule Applied

We're focusing on the 20% of improvements that will give us 80% of the value:

1. Fix actual bugs (not theoretical issues)
2. Add context where it matters (not everywhere)
3. Improve errors simply (not comprehensively)
4. Remove obvious duplication (not all duplication)
5. Complete missing features (not redesign everything)

## Definition of Done (Simplified)

- [ ] No compilation errors
- [ ] Critical tests passing
- [ ] Context support for long operations
- [ ] Basic error wrapping
- [ ] Missing features implemented
- [ ] Measured performance (optimize only if slow)
- [ ] Basic documentation updated

## Implementation Philosophy

1. **Make it work** - We're already at 82%, so mostly done
2. **Make it right** - Small, targeted improvements
3. **Make it fast** - Only if measurements show it's slow

**Remember**:

- Perfect is the enemy of good
- Ship working code, not perfect code
- Complexity is a bug, not a feature

---

*This pragmatic plan delivers a production-ready parser in 2 weeks by focusing on what matters and skipping over-engineering.*
