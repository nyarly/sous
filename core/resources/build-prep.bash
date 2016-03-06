#!/usr/bin/env bash
. /etc/profile
PATH="$PATH:/usr/local/bin"
log() { echo "$@" >&2; }
die() { log "$@"; exit 1; }

REPO_MNT="/mnt/repo"

set -eu

[ ! -d "$REPO_MNT" ] && die "You must mount your repository's root dir to $REPO_MNT using docker run -v \$PWD:$REPO_MNT"

# Copy all the git indexed and new unignored files to /build
# (this is an analogue of docker's own 'send-context')
# The --exclude-standard and --others flags together ensure that
# new unindexed, unignored files get copied along with everything in
# the index.
cd "$REPO_MNT"
git ls-files --exclude-standard --others --cached | while read f; do
	# Ensure the dir heirarcy exists, as cp is unable to do this in a cross-platform way
	if ! [[ $(dirname $f) == "$f" ]]; then
		DIR="/build/$(dirname $f)"
		[ -d "$DIR" ] || mkdir -p "$DIR"
	fi
	cp -f "$f" "/build/$f" || log "WARNING: Unable to copy $f - you may have deleted it but not yet committed the delete."
done

