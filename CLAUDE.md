# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hermes is a high-performance web content extraction library inspired by Postlight Parser that transforms web pages into clean, structured data. It extracts article content, titles, authors, dates, images, and more from any URL using site-specific custom parsers and generic fallback extraction.

## Development Commands

### Build Commands
- `yarn build` - Full build with linting and testing (includes Node.js target)
- `yarn build:ci` - CI build without linting
- `yarn build:web` - Build for web/browser target with linting
- `yarn build:esm` - Build ES module version with linting
- `yarn build:web:ci` - Web build for CI without linting
- `yarn build:esm:ci` - ESM build for CI without linting
- `yarn release` - Build both Node and web versions

### Testing Commands
- `yarn test` - Run all tests (Node.js and web)
- `yarn test:node` - Run Jest tests for Node.js (outputs test-output.json)
- `yarn test:web` - Run Karma tests for web/browser
- `yarn test:build` - Test that build artifacts work correctly
- `yarn test:build:web` - Test web build
- `yarn test:build:esm` - Test ESM build
- `yarn watch:test` - Run Jest in watch mode
- `yarn watch:test <domain>` - Watch tests for specific custom parser (e.g., `yarn watch:test www.newyorker.com`)

### Linting and Code Quality
- `yarn lint` - Run ESLint with auto-fix
- `yarn lint:ci` - Run linting for CI (includes remark for markdown)
- `yarn lint-fix-quiet` - Run ESLint with quiet auto-fix

### Custom Parser Development
- `yarn generate-parser` - Interactive generator for new custom parsers
- `./preview <url>` - Preview extraction results for a URL (generates HTML and JSON)

## Architecture Overview

### Core Components

**Main Entry Point (`src/mercury.js`)**
- Primary Parser class with `parse()` method
- Handles URL validation, resource fetching, and content extraction orchestration
- Supports multiple content formats (HTML, Markdown, text)
- Manages custom headers, pre-fetched HTML, and pagination

**Resource Layer (`src/resource/`)**
- Handles HTTP requests and HTML fetching
- DOM manipulation and normalization utilities
- Encoding detection and lazy image conversion

**Extractor System (`src/extractors/`)**
- **Custom Extractors** (`src/extractors/custom/`) - Site-specific parsers (150+ sites)
- **Generic Extractors** (`src/extractors/generic/`) - Fallback content extraction
- **Root Extractor** - Coordinates extraction process and field selection

**Content Processing (`src/cleaners/`, `src/utils/`)**
- Field-specific cleaners for title, author, date, content, etc.
- DOM utilities for content scoring, cleaning, and transformation
- Text processing utilities for excerpts, encoding, and normalization

### Custom Parser Structure

Custom parsers are the heart of site-specific extraction. Each parser exports:

```javascript
export const SiteExtractor = {
  domain: 'example.com',
  title: { selectors: ['h1.headline'] },
  author: { selectors: ['.byline'] },
  content: { 
    selectors: ['.article-body'],
    clean: ['.ads', '.related'],
    transforms: { 'h1': 'h2' }
  },
  date_published: { selectors: [['time', 'datetime']] },
  extend: {
    category: { selectors: ['.tags a'], allowMultiple: true }
  }
};
```

### Build System

- **Rollup** for bundling with Babel transpilation
- Multiple build targets: Node.js (CJS), Web (UMD), ESM
- **Jest** for Node.js testing, **Karma** for browser testing
- **ESLint** with Airbnb config and Prettier formatting

### Testing Strategy

- **Fixture-based testing** - HTML snapshots in `fixtures/` directory
- **Custom parser tests** - Each parser has comprehensive test coverage
- **Build verification** - Tests ensure compiled artifacts work correctly
- **Cross-platform testing** - Node.js and browser environments

## Key Development Workflows

### Adding a Custom Parser

1. Run `yarn generate-parser` and provide article URL
2. Edit generated parser in `src/extractors/custom/[domain]/index.js`
3. Use browser dev tools on fixture HTML to find CSS selectors
4. Run `yarn watch:test [domain]` to iterate on failing tests
5. Use `./preview <url>` to verify content extraction quality

### Parser Testing

Custom parsers use fixture HTML files for consistent testing. Tests verify:
- Correct extractor selection
- Field extraction accuracy (title, author, content, etc.)
- Content cleaning and transformation
- Custom field extensions

### Content Scoring Algorithm

The generic extractor uses a sophisticated scoring system:
- **Node scoring** based on content density, text length, and HTML structure
- **Sibling merging** to combine related content blocks
- **Link density analysis** to avoid navigation-heavy content
- **Candidate selection** from top-scoring content nodes

## File Organization

- `src/extractors/custom/` - Site-specific parsers (organized by domain)
- `src/extractors/generic/` - Fallback extraction algorithms
- `src/cleaners/` - Field-specific content cleaning
- `src/utils/dom/` - DOM manipulation utilities
- `src/utils/text/` - Text processing utilities
- `fixtures/` - HTML test fixtures for parser validation
- `scripts/` - Build tools and development utilities

## Dependencies

- **cheerio** - Server-side jQuery-like DOM manipulation
- **turndown** - HTML to Markdown conversion
- **moment** - Date parsing and formatting
- **difflib** - Content comparison for testing
- **rollup** - Module bundling
- **jest/karma** - Testing frameworks