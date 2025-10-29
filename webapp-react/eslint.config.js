// @ts-check

import eslint from "@eslint/js";
import query from "@tanstack/eslint-plugin-query";
import eslintConfigPrettier from "eslint-config-prettier/flat";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config({
  files: ["src/**/*.{js,jsx,mjs,cjs,ts,tsx}"],
  extends: [
    eslint.configs.recommended,
    tseslint.configs.recommendedTypeChecked,
    // tseslint.configs.recommended,
    react.configs.flat.recommended,
    react.configs.flat["jsx-runtime"],
    query.configs["flat/recommended"],
    eslintConfigPrettier,
  ],
  plugins: {
    "react-hooks": reactHooks,
  },
  languageOptions: {
    parserOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      ecmaFeatures: {
        jsx: true,
      },
      project: "./tsconfig.json",
      tsconfigRootDir: import.meta.dirname,
    },
    globals: {
      ...globals.browser,
    },
  },
  rules: {
    "@typescript-eslint/no-unused-expressions": 0,
    "react-hooks/rules-of-hooks": "error",
    "react-hooks/exhaustive-deps": "error",
  },
});
