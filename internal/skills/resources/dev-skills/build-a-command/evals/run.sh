#!/usr/bin/env bash
# Runs the build-a-command skill's evals. Each eval copies this checkout to a
# scratch directory, gives a fresh headless Claude Code session only its
# prompt, with the skill pointer and the osapi CLI available as in any checkout, and
# then checks the result. The session never sees the checks.
#
#   internal/skills/resources/dev-skills/build-a-command/evals/run.sh            # every eval
#   internal/skills/resources/dev-skills/build-a-command/evals/run.sh greet      # one
#
# BUDGET caps each session's spend in US dollars (default 15). KEEP=1 keeps
# the scratch copies. Results are written to $OUT (default a temp directory):
# each eval's session transcript as JSON and a results.txt of PASS and FAIL.
set -uo pipefail

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../../../../.." && pwd)
out=${OUT:-$(mktemp -d -t osapi-skill-evals)}
budget=${BUDGET:-15}
ids=("$@")
if [ ${#ids[@]} -eq 0 ]; then
	mapfile -t ids < <(python3 -c 'import json,sys; [print(e["id"]) for e in json.load(open(sys.argv[1]))["evals"]]' "$here/evals.json")
fi

field() { python3 -c 'import json,sys; e=[e for e in json.load(open(sys.argv[1]))["evals"] if e["id"]==sys.argv[2]][0]; print(e.get(sys.argv[3],""))' "$here/evals.json" "$1" "$2"; }

result() { # eval check status
	printf '%-8s %-5s %s\n' "$1" "$3" "$2" | tee -a "$out/results.txt"
}

check() { # eval name command...
	local id=$1 name=$2
	shift 2
	if (cd "$work" && eval "$@") >>"$out/$id.checks.log" 2>&1; then
		result "$id" "$name" PASS
	else
		result "$id" "$name" FAIL
	fi
}

# One snapshot of the checkout, taken before any session starts, so every
# eval begins from the same tree even if the checkout changes while they run.
snapshot="$out/snapshot"
mkdir -p "$snapshot"
rsync -a --exclude .git --exclude docs/build --exclude .worktrees "$repo/" "$snapshot/"

for id in "${ids[@]}"; do
	prompt=$(field "$id" prompt)
	g=$(field "$id" group)
	work="$out/$id/osapi"
	mkdir -p "$work"
	rsync -a "$snapshot/" "$work/"
	(cd "$work" && git init -q && git add -A && git -c user.name=eval -c user.email=eval@localhost commit -qm baseline)

	echo "== $id: session started $(date +%H:%M:%S), budget \$$budget" | tee -a "$out/results.txt"
	(cd "$work" && claude -p "$prompt" \
		--permission-mode acceptEdits \
		--allowedTools "Bash(go:*)" "Bash(mise:*)" "Bash(just:*)" "Bash(gofmt:*)" "Bash(git status:*)" "Bash(git diff:*)" "Bash(ls:*)" "Bash(grep:*)" "Bash(find:*)" "Bash(cat:*)" "Bash(sed:*)" \
		--max-budget-usd "$budget" \
		--output-format json) >"$out/$id.session.json" 2>"$out/$id.session.err"
	python3 -c 'import json,sys; r=json.load(open(sys.argv[1])); print("   turns=%s cost=$%.2f seconds=%d" % (r.get("num_turns"), r.get("total_cost_usd",0), r.get("duration_ms",0)/1000))' "$out/$id.session.json" 2>/dev/null | tee -a "$out/results.txt"

	answer="python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get(\"result\",\"\"))' '$out/$id.session.json'"
	if [ "$(field "$id" kind)" = ask ]; then
		check "$id" "nothing in the checkout changed" "[ -z \"\$(git status --porcelain --untracked-files=all)\" ]"
		check "$id" "the answer asks the user at least two questions" "[ \"\$($answer | grep -c '?')\" -ge 2 ]"
	elif [ "$(field "$id" kind)" = refuse ]; then
		expects=$(field "$id" expects)
		check "$id" "nothing in the checkout changed" "[ -z \"\$(git status --porcelain --untracked-files=all)\" ]"
		check "$id" "the answer says that it is a$( [ "$expects" = endpoint ] && echo 'n') $expects" "$answer | grep -qi '$expects'"
	else
		check "$id" "dsl command validate passes" "go run . dsl command validate"
		check "$id" "the group is registered under its name" "go run . dsl command get --name $g --json"
		check "$id" "its package tests pass with 100% coverage" "go test -cover ./internal/cli/commands/$g/ | tee /dev/stderr | grep -q 'coverage: 100.0%'"
		check "$id" "every test in internal/cli/commands/all passes" "go test -count=1 ./internal/cli/commands/all/"
		check "$id" "generated files are current" "before=\$(git status --porcelain); go generate ./internal/cli/commands/all && mise exec -- just cli-docs-generate && [ \"\$before\" = \"\$(git status --porcelain)\" ]"
		check "$id" "lint reports no issues in the group" "go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .golangci.yml ./internal/cli/commands/$g/..."
		check "$id" "nothing outside the group, its registration, its wiring and generated files changed" \
			"! git status --porcelain --untracked-files=all | awk '{print \$2}' | grep -vE '^(internal/cli/commands/$g/|internal/cli/commands/all/|cmd/|docs/docs/sidebar/reference/|docs/docs/sidebar/usage/cli/)'"
		[ "$(field "$id" refusals)" = "True" ] && check "$id" "at least one refusal is declared with Fails" "grep -rqE '(^|[.[:space:]])Fails\\(' internal/cli/commands/$g/"
	fi

	(cd "$work" && git status --porcelain --untracked-files=all) >"$out/$id.changed.txt"
	[ "${KEEP:-}" = 1 ] || rm -rf "$out/$id"
done

[ "${KEEP:-}" = 1 ] || rm -rf "$snapshot"
echo "results: $out/results.txt"
