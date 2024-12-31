/** @type {import("eslint").Linter.Config} */
module.exports = {
  root: true,
  extends: ['@northern.tech/eslint-config/react.js', 'plugin:import/typescript'],
  ignorePatterns: ['node_modules/', 'dist/', '**/api/types/', 'tests/licenses'],
  globals: {
    mender_environment: 'readonly',
    NodeJS: true
  },
  overrides: [
    {
      files: ['tests/e2e_tests/**/*'],
      rules: {
        'react-hooks/rules-of-hooks': 'off'
      }
    }
  ]
};
