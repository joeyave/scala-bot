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
  settings: {
    react: {
      version: "detect",
    },
  },
  rules: {
    "@typescript-eslint/no-unused-expressions": 0,
    "react-hooks/rules-of-hooks": "error",
    "react-hooks/exhaustive-deps": "error",
    "@typescript-eslint/no-explicit-any": "off",
    "@typescript-eslint/no-unsafe-assignment": "off",
    "@typescript-eslint/no-unsafe-member-access": "off",
    "@typescript-eslint/no-unsafe-argument": "off",
    "@typescript-eslint/no-unnecessary-type-assertion": "off",
    "@typescript-eslint/no-redundant-type-constituents": "off",
    "no-useless-assignment": "off",
  },
});
