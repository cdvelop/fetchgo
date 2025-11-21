# Main Prompt: Refactor and Test fetchgo Library

## Goal

Refactor the `fetchgo` library to use an automatic codec detection system and increase test coverage to over 90%.

## Instructions for Jules

Hi Jules! This project is divided into two main parts:
1. **Architecture Refactor**: Implement a new, simplified API with automatic codec selection.
2. **Test Implementation**: Write comprehensive tests for the new architecture.

**Please follow the documents in order.** Start with the architecture, then implement the tests.

### Part 1: Architecture Refactor

**Summary**: The goal is to remove manual encoder configuration and replace it with an automatic system based on `Content-Type` headers.

**Required Reading**:
- `ARCHITECTURE.md`: Detailed explanation of the new architecture.
- `IMPLEMENTATION_GUIDE.md`: Step-by-step implementation checklist.

**Key Changes**:
- `Fetchgo` struct becomes a codec manager.
- `New()` creates a zero-config instance.
- `NewClient(timeoutMS)` creates clients WITHOUT a `baseURL` parameter.
- **CRITICAL**: All URLs must be absolute (e.g., `https://api.example.com/users`).
- A single client can make requests to multiple domains.
- **Two explicit methods** to avoid confusion:
  - `SendJSON()` → Encodes as JSON (`application/json`)
  - `SendBinary()` → Encodes with TinyBin (`application/octet-stream`)
- User chooses which encoding to use (explicit, no guessing).
- CORS is configured via a simple `SetCORS()` method.
- Simple API: `SendJSON()`, `SendBinary()`, and `SetHeader()` methods.

### Part 2: Test Implementation

**Summary**: After implementing the new architecture, write tests to achieve 90%+ coverage.

**Required Reading**:
- `TESTING_PLAN.md`: Detailed testing strategy and checklist.

**Key Areas to Test**:
- New `Fetchgo` instance and client creation (NO baseURL).
- Automatic codec selection logic.
- Platform-specific codecs (stdlib vs. WASM).
- Simple encoder system (JSON for structs, raw pass-through for []byte).
- CORS configuration in WASM.
- **Multi-domain support**: Single client making requests to different domains.

**IMPORTANT**:
- **Do NOT run WASM tests yourself.** Your environment often fails. Just write the code. We will handle the browser testing.
- **Follow the implementation guide first.** The tests depend on the new architecture.
- **Commit incrementally.** Don't wait until everything is finished.
- **Note**: For multi-domain testing, you will need to run TWO test servers on different ports (8080 and 9090).

Good luck! 🚀
