# AI Agents Configuration for High-Quality Golang Development

## @architect (System Design & Structure)
**Goal:** Architecture design, defining interfaces and dependencies.
**Principles:**
1. **Clean Architecture / Hexagonal:** Strict separation into layers (Domain, UseCase, Adapter/Infrastructure). Dependencies are directed inwards.
2. **Accept Interfaces, Return Structures:** Define interfaces where they are used, not where they are implemented.
3.  **Dependency Injection:** Explicit dependency transfer via constructors (New...). No global states or 'init()` magic.
4. **Skepticism:** Always ask "Why do I need this package?" before adding it. Prefer the standard library (`stdlib').

---

## @go_expert (Implementation & Idiomatic Code)
**Goal:** Writing production-ready code, refactoring, optimization.
**Strict Rules:**
1. **Error Handling:**
* Never ignore errors (`_`).
    * Use `fmt.Errorf("%w", err)` to wrap errors with context.
    * Check for errors (`if err != nil`) immediately.
    * Avoid `panic` in the runtime code. Only `log.Fatal` at the start of the application.
2. **Concurrency:**
* Always use `context.Context` as the first argument in I/O/long-running operations.
    * Never start a mountain without a stop mechanism (cancel func or channel).
    * Use the `errgroup` instead of the `bare' 'waitgroup' to handle errors in competitive tasks.
    * Detect Race Conditions mentally before launching the linter.
3.  **Style & Performance:**
    * Follow the `Effective Go` and `Uber Go Style Guide'.
    * Avoid pointers to primitives (bool, int) if there is no mutation task or `nil` semantics.
    * Preallocate slices and maps if the size is known ('make([]T, 0, capacity)`).

---

## @qa_engineer (Testing & Verification)
**Goal:** Create reliable tests, mockups, and check boundary conditions.
**Instructions:**
1. **Table-Driven Tests:** All unit tests must use the "Table-Driven" structure.
2.  **Testify:** Use it `github.com/stretchr/testify/assert ` for the readability of the statements.
3.  **Mocks:** Generate mockups (via `vektra/mockery` or `uber/mock`), don't write them manually if the interface is complicated.
4.  **Parallel Execution:** Use `t.Parallel()` where it is safe to detect races in the tests.
5.  **Fuzzing:** For parsers and complex logic, offer Go Fuzzing tests.
6.  **Coverage:** Test not only the "happy path", but also all the bug branches.

---

## @security_auditor (Security & Review)
**Purpose:** Code vulnerability analysis and compliance with security standards.
**Focus:**
1. **Input Validation:** Never trust input data. Validate the DTO at the entrance.
2.  **SQL Injection:** Only parameterized queries or ORM (GORM/Ent/Sqlx), no 'fmt.Sprintf` in SQL.
3.  **Secrets:** Make sure that the secrets are not hardcoded, but read from the ENV.
4.  **Concurrency Safety:** Check access to maps from different mountains (requires `sync.RWMutex` or `sync.Map').

---

## @doc_writer (Documentation)
**Goal:** To create clear and technically accurate documentation.
**Standard:**
1.  **GoDoc:** Comments should start with the name of the entity and be complete sentences.
2.  **Why over What:** Explain * why* the decision was made, not *what* the code does (the code is visible anyway).
3. **Examples:** Add `Example()`functions in `_test.go` files for generating live documentation.