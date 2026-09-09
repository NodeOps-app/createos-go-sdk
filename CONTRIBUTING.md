# Contributing

Thank you for contributing to the CreateOS Go SDK.

## Setup

The repository pins Go and golangci-lint with asdf:

```sh
asdf install
make install-hooks
make check
make test-race
```

`make install-hooks` configures Git to use the repository's version-controlled
hooks. Run it once after cloning.

## Commit convention

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```text
<type>(<optional-scope>): <subject>
```

Accepted types are `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`,
and `perf`. Write the subject in imperative mood, keep it at 50 characters or
fewer, start it with a lowercase letter, and do not end it with a period.

Examples:

```text
feat(sandbox): add disk attachment
fix(transport): honor retry-after header
docs(readme): add network example
test(processes): cover output replay
chore: prepare v0.0.1 release
```

For an incompatible public API change, add `!` before the colon and explain the
change in a `BREAKING CHANGE:` commit footer:

```text
feat(sandbox)!: change create response
```

The local `commit-msg` hook validates the machine-checkable rules. Pull-request
CI validates every commit in the branch, so the same rules apply even when a
contributor has not installed the hook.

## Before opening a pull request

Run:

```sh
make check
make test-race
```

Add or update tests for behavior changes and update public GoDoc and README
examples when the public API changes.
