#!/usr/bin/env bash
. /etc/profile
PATH="$PATH:/usr/local/bin"
log() { echo "$@" >&2; }
die() { log "$@"; exit 1; }

BUILD_COMMAND="$@"
[ -z "$BUILD_COMMAND" ] && die "You must pass a build command to execute, e.g. 'npm install'"

[ -z "$REPO_WORKDIR" ] && die "You must set REPO_WORKDIR, use / for repo root."

if [ -z "$ARTIFACT_NAME" ]; then
	log "WARNING: ARTIFACT_NAME env var nor set; using 'artifact'"
	ARTIFACT_NAME=artifact
fi

set -eu

[ ! -d /repo ] && die "You must mount your working dir to /repo using docker run -v \$PWD:/repo"
[ ! -d /artifacts ] && die "You must mount your artifact output dir to /artifacts using \$somepath:/artifacts"

# Ensure the build dir exists
[ -d /build ] || mkdir /build

# Copy all the git indexed and new unignored files to /build
# (this is an analogue of docker's own 'send-context')
# The --exclude-standard and --others flags together ensure that
# new unindexed, unignored files get copied along with everything in
# the index.
cd "/repo$REPO_WORKDIR"
git ls-files --exclude-standard --others --cached | while read f; do
	# Ensure the dir heirarcy exists, as cp is unable to do this in a cross-platform way
	if ! [[ $(dirname $f) == "$f" ]]; then
		DIR="/build/$(dirname $f)"
		[ -d "$DIR" ] || mkdir -p "$DIR"
	fi
	cp -f "$f" "/build/$f" || log "WARNING: Unable to copy $f - you may have deleted it but not yet committed the delete."
done

# Set working directory to /build; the passed "BUILD_COMMAND" is executed in here
cd /build

# Execute the build command inside the isolated /build dir
eval "$BUILD_COMMAND"

ARTIFACT_FILENAME="$ARTIFACT_NAME.tar.gz"
ARTIFACT_PATH="/artifacts/$ARTIFACT_FILENAME"

# All done, zip up the results and plonk in /artifacts
if [ -f "$ARTIFACT_PATH" ]; then
	log "Deleting old artifact '$ARTIFACT_FILENAME'"
	rm "$ARTIFACT_PATH"
fi

tar -czf "$ARTIFACT_PATH" . || die "Failed to create tarball."
if [ ! -z "$ARTIFACT_OWNER" ]; then
	log "Setting artifact owner to $ARTIFACT_OWNER"
	chown "$ARTIFACT_OWNER" "$ARTIFACT_PATH" || \
		die "Failed to set owner of $ARTIFACT_PATH to $ARTIFACT_OWNER"
else
	log "No artifact owner specificied."
fi

log "Successfully created $ARTIFACT_FILENAME"

