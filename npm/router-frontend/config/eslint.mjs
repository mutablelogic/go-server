import { FlatCompat } from '@eslint/eslintrc';
import globals from 'globals';
import {dirname } from 'path';
import { fileURLToPath } from 'url';
import js from '@eslint/js';

// mimic CommonJS variables -- not needed if using CommonJS
const _filename = fileURLToPath(import.meta.url);
const _dirname = dirname(_filename);
const compat = new FlatCompat({
  baseDirectory: _dirname,
});

export default [
  js.configs.recommended,
  {
    languageOptions: {
      globals: globals.browser,
      ecmaVersion: 'latest',
      sourceType: 'module',
    },
  },
  ...compat.extends('airbnb-base'),
  {
    rules: {
      'import/prefer-default-export': 'off',
    },
  }
];
