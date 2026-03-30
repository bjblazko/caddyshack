# ADR-003: Vanilla JS Frontend — No Framework

**Date:** 2026-03-29
**Status:** Accepted

## Context

Modern web UIs are commonly built with React, Vue, Svelte, or similar frameworks. These provide component models, state management, and reactivity, but require a build step (npm, bundler, transpiler) and add hundreds of kilobytes of runtime code.

CaddyShack's UI is a single-page dashboard with a limited number of interactive states: file uploaded / not uploaded, host selection, and traffic filter selection. There is no complex nested component tree or high-frequency state changes.

## Decision

The frontend is written in vanilla HTML5, CSS3, and JavaScript (ES2020+). There is no build step, no npm, no bundler, and no transpiler. D3.js is used for the world map only and is the single external dependency (vendored locally).

DOM manipulation, event handling, and state (current host, current filter) are managed directly in `app.js` using standard Web APIs.

## Consequences

**Positive:**
- No build toolchain to install, configure, or maintain
- Frontend can be edited and tested by opening `index.html` directly in a browser (with a local server for API calls)
- Zero npm dependency supply-chain risk
- Low barrier to contribution: any developer familiar with the Web platform can read and modify the code
- Smaller total JavaScript payload than a framework-based app

**Negative:**
- No component model: code organization relies on naming conventions and namespaced module objects (`Charts`, `WorldMap`) rather than framework abstractions
- State management is imperative rather than declarative — more verbose for complex interactions
- No hot-module reload or similar developer-experience tooling
