import reactConfig from '@northern.tech/eslint-config/react.js';
import globals from 'globals';

const globalsConfig = {
  ...globals.commonjs,
  ...globals.node,
  ...globals.vitest,
  ...globals.denoBuiltin,
  mender_environment: 'readonly'
};

export default [
  ...reactConfig,
  { ignores: ['node_modules/', 'dist/', '.yalc/'] },
  {
    languageOptions: { globals: globalsConfig },
    rules: {
      'react-hooks/refs': 'off',
      'react-hooks/set-state-in-effect': 'off'
    }
  },
  {
    files: ['tests/e2e_tests/**/*'],
    languageOptions: { globals: globalsConfig },
    rules: {
      'react-hooks/rules-of-hooks': 'off'
    }
  }
];
