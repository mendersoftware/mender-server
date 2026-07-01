# Contributing to the Mender Frontend

First off, thank you for considering contributing to the Mender frontend! We
welcome contributions from the community. This document provides guidelines
specifically for contributing to the code within the `frontend` directory.

Please also read the main project contribution guide at
[`../CONTRIBUTING.md`](../CONTRIBUTING.md) for general information about issues,
pull requests, sign-off, commit structure, and the code of conduct.

## Getting Started

Before you start coding, it's usually best to:

1. Find an existing issue or create a new one in our [issue tracker](https://northerntech.atlassian.net/jira/software/c/projects/MEN/list) to discuss the change you want to make.
2. Ensure you have the development environment set up correctly as described in the frontend [README](./README.md).

## Frontend Development Guidelines

To ensure consistency, maintainability, and quality in our frontend codebase, we ask that you adhere to the following guidelines:

### 1. Code Style & Formatting

- **Prettier:** We use Prettier to enforce a consistent code style. Our configuration is defined in [`.prettierrc.mjs`](./.prettierrc.mjs).
- **Enforcement:** Code style is automatically checked in our CI pipeline. Pull requests with formatting issues will fail checks.
- **How to Apply:** Before committing your changes, please run the formatting command:
  ```bash
  npm run lint-fix
  ```
  Consider configuring your editor to format on save using the project's Prettier configuration.

### 2. Naming Conventions

- **Descriptive Names:** Variables, functions, components, and file names should be descriptive and clearly convey their purpose. Avoid abbreviations unless they are widely understood (e.g., `id`, `url`, `http`).
- **Consistency:**
  - **Variables & Functions:** Use `camelCase` (e.g., `isLoading`, `fetchDeviceDetails`).
  - **React Components:** Use `PascalCase` for component names and filenames (e.g., `DeviceList.js`, `function DeviceList(...)`).
  - **Constants:** Use `UPPER_SNAKE_CASE` for true constants (e.g., `MAX_LOGIN_ATTEMPTS`).
  - **CSS/SCSS classes:** Use kebab-case (e.g., `.device-list-item`). Follow existing patterns where applicable.

### 3. Testing

We aim for a high degree of confidence in our frontend application through testing. Our testing philosophy is heavily inspired by Kent C. Dodds' principles, particularly the Testing Trophy and the practices promoted by React Testing Library.

- **Testing Trophy Focus:**
  - **(Few) End-to-End Tests:** Cover critical user flows using tools like Cypress (if applicable). These are valuable but slower and more brittle.
  - **(Many) Integration Tests:** **This is where the bulk of our testing effort should be.** Use React Testing Library to test components within the context they are used, verifying interactions and rendered output from a user's perspective. Test the integration between several units/components.
  - **(Some) Unit Tests:** Use for pure functions, complex logic, utility functions, or hooks that can be tested in isolation without a UI. Avoid unit-testing component implementation details.
- **React Testing Library (RTL) Best Practices:**
  - **Query Like a User:** Prioritize querying elements by accessible attributes (roles, labels, text content) using queries like `getByRole`, `getByLabelText`, `getByText`. Use `data-testid` sparingly as a last resort.
  - **Test Behavior, Not Implementation:** Focus on *what* the component does for the user, not *how* it internally achieves it. Avoid testing component state or instance methods directly.
  - **Accessibility:** Writing tests using accessible queries helps ensure our application is usable by everyone.
- **Avoid Mocking When Possible:**
  - **Why:** Over-mocking leads to tests that are tightly coupled to implementation details. If you refactor the component's internals without changing its behavior, a heavily mocked test might break unnecessarily. Real components interacting provide more confidence.
  - **Prefer Integration:** Instead of mocking child components, render them and test the integrated behavior.
  - **When Mocks Are Okay:**
    - **Network Requests:** Mock API calls using tools like Mock Service Worker (MSW) or similar established patterns within the project. Do not mock `fetch` or `axios` directly if a higher-level abstraction is available.
    - **External Dependencies:** Browser APIs not available in the test environment (e.g., `localStorage`, `matchMedia`) might need mocking.
    - **True Unit Tests:** When isolating a complex algorithm or utility function.
- **Test Location & Execution:**
  - Tests typically reside alongside the code using `.test.ts` extensions. Follow the existing project structure.
  - Run the tests using:
    ```bash
    npm run test
    ```
  - Ensure all tests pass before submitting your pull request.

## Pull Request Process

1. Ensure your code adheres to the guidelines above (formatting, naming, testing).
2. Rebase your branch onto the latest `main` (or the target branch) before submitting.
3. Follow the sign-off and commit-message rules in the [main contribution guide](../CONTRIBUTING.md): commits use the Conventional Commits format (scope `gui` for frontend changes) and must be signed off, as enforced by [`commitlint.config.js`](../commitlint.config.js).
4. Provide a clear and descriptive title and description for your pull request.
   - Link the relevant issue (e.g., `Ticket: MEN-123`).
   - Explain *what* changes were made and *why*.
   - Include screenshots or GIFs for UI changes.
5. Ensure all CI checks pass.
6. Be prepared to discuss your changes and make adjustments based on reviewer feedback.
