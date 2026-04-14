module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      ["feat", "fix", "docs", "refactor", "test", "chore", "revert"],
    ],
    "scope-empty": [2, "never"],
    "scope-enum": [2, "always", ["core", "ai", "cross"]],
  },
};
