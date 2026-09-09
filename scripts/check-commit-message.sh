#!/usr/bin/env bash

set -euo pipefail

usage() {
	echo "usage: $0 --file <commit-message-file> | --subject <subject> | --range <git-range>" >&2
	exit 2
}

validate_subject() {
	local subject=$1
	local pattern='^(feat|fix|docs|style|refactor|test|chore|perf)(\([a-z0-9][a-z0-9._/-]*\))?(!)?: (.+)$'

	if [[ ! $subject =~ $pattern ]]; then
		echo "invalid commit subject: $subject" >&2
		echo "expected: <type>(<optional-scope>): <subject>" >&2
		echo "types: feat, fix, docs, style, refactor, test, chore, perf" >&2
		return 1
	fi

	local description=${BASH_REMATCH[4]}
	if (( ${#description} > 50 )); then
		echo "commit subject text exceeds 50 characters: $description" >&2
		return 1
	fi
	if [[ $description == .* ]]; then
		echo "commit subject text must not start with a period: $description" >&2
		return 1
	fi
	if [[ $description == *\. ]]; then
		echo "commit subject text must not end with a period: $description" >&2
		return 1
	fi
	if [[ $description =~ ^[A-Z] ]]; then
		echo "commit subject text must start with a lowercase letter: $description" >&2
		return 1
	fi
}

if (( $# != 2 )); then
	usage
fi

case $1 in
	--file)
		IFS= read -r subject < "$2"
		validate_subject "$subject"
		;;
	--subject)
		validate_subject "$2"
		;;
	--range)
		failed=0
		while IFS= read -r subject; do
			if ! validate_subject "$subject"; then
				failed=1
			fi
		done < <(git log --format=%s "$2")
		exit "$failed"
		;;
	*)
		usage
		;;
esac
