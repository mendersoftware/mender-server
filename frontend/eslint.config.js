import reactConfig from '@northern.tech/eslint-config/react.js';
import globals from 'globals';

const globalsConfig = {
  ...globals.commonjs,
  ...globals.node,
  ...globals.vitest,
  mender_environment: 'readonly'
};

export default [
  ...reactConfig,
  { ignores: ['node_modules/', 'dist/', '**/api/types/**', 'tests/licenses'] },
  {
    languageOptions: { globals: globalsConfig }
  },
  {
    files: ['tests/e2e_tests/**/*'],
    languageOptions: { globals: globalsConfig },
    rules: {
      'react-hooks/rules-of-hooks': 'off'
    }
  }
];
