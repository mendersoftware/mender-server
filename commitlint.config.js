const commonScopes = [
  'create-artifact-worker',
  'deployments',
  'deviceauth',
  'deviceconfig',
  'deviceconnect',
  'inventory',
  'iot-manager',
  'reporting',
  'useradm',
  'workflows',
  'gui',
  null
];

module.exports = {
  extends: ['@commitlint/config-conventional'],
  plugins: ['selective-scope'],
  rules: {
    'body-max-line-length': [1, 'always', 100],
    'selective-scope': [
      2,
      'always',
      {
        feat: commonScopes,
        fix: commonScopes,
        test: commonScopes,
        perf: [],  // scope is not allowed
        ci: []     // scope is not allowed
      }
    ],
    'subject-case': [1, 'always', ['lower-case', 'sentence-case']],
  },

  helpUrl: `
  Commit messages must follow conventional commit format:
  https://www.conventionalcommits.org/en/v1.0.0/#summary
      type(optional-scope): subject

      [optional body]
  * To bypass pre-commit hooks run 'git commit --no-verify'
  >>> Use "npm run commit" for interactive prompt. <<<
  `
};

