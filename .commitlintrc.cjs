module.exports = {
    extends: ["@commitlint/config-conventional"],
    rules: {
      "subject-case": [2, "always", ["sentence-case", "lower-case", "start-case"]],
      "type-enum": [
        2,
        "always",
        [
          "feat", "fix", "docs", "style", "refactor",
          "perf", "test", "build", "ci", "chore", "revert"
        ]
      ],
      "scope-empty": [0, "never"]
    }
  };
  