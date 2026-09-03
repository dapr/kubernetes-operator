#!/bin/sh
#
# Run govulncheck and fail only on vulnerabilities that the project can act on.
#
# govulncheck has no ignore flag, so this script reads the JSON report and
# filters a small allowlist. Every entry below has no fixed version. Review the
# list whenever a dependency changes. Remove an entry as soon as a fix ships.
#
# Usage: govulncheck.sh <govulncheck command...>

set -eu

# GO-2026-5064  containerd  CRI checkpoint restore CDI annotation smuggling
# GO-2026-5338  containerd  CRI checkpoint import allows local image tag poisoning
# GO-2026-5622  containerd  Arbitrary host CRI log file read through a symlink
#
#   All three are fixed only in github.com/containerd/containerd/v2. The v1
#   line is affected from version 0 and has no fix event. The operator reaches
#   it through helm.sh/helm/v3/pkg/registry, which uses containerd/remotes for
#   OCI transport. The operator does not call the CRI checkpoint API.
#
# GO-2026-5932  golang.org/x/crypto  openpgp is unmaintained and unsafe by design
#
#   The advisory has no fixed version. The operator reaches x/crypto through
#   helm.sh/helm/v3/pkg/action and Masterminds/sprig/v3.
ALLOW='GO-2026-5064 GO-2026-5338 GO-2026-5622 GO-2026-5932'

REPORT=$(mktemp)
trap 'rm -f "$REPORT"' EXIT

echo "run govulncheck"
rc=0
"$@" -format json ./... > "$REPORT" || rc=$?

# govulncheck exits 0 with no findings and 3 with findings. Any other code is
# a scanner or build error, so fail rather than report a clean run.
if [ "$rc" -ne 0 ] && [ "$rc" -ne 3 ]; then
    echo "govulncheck failed with exit code $rc" >&2
    exit 1
fi

if [ ! -s "$REPORT" ]; then
    echo "govulncheck produced no report" >&2
    exit 1
fi

# A vulnerability counts as called when at least one finding names a function.
CALLED=$(jq -rs '
    [ .[]
      | select(has("finding"))
      | .finding
      | select([ .trace[]? | select(has("function")) ] | length > 0)
      | .osv
    ] | unique | .[]
' "$REPORT")

allowed=''
blocking=''
for id in $CALLED; do
    case " $ALLOW " in
        *" $id "*) allowed="$allowed $id" ;;
        *)         blocking="$blocking $id" ;;
    esac
done

if [ -n "$allowed" ]; then
    echo "allowed, no fix available:$allowed"
fi

if [ -n "$blocking" ]; then
    echo ""
    echo "govulncheck found vulnerabilities that are not allowlisted:$blocking"
    echo ""
    "$@" ./... || true
    exit 1
fi

echo "no actionable vulnerabilities"
