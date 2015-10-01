#!/usr/bin/env bash
log() { echo "$@" 1>&2; }
die() { log "$@"; exit 1; }
if [ -z "$1" ]; then
	die "usage: $0 self | all | TARGET..."
fi
REQUESTED_TARGETS=($@)
STANDARD_TARGETS=(linux/amd64 darwin/amd64)
if [[ "$1" == "all" ]]; then
	REQUESTED_TARGETS=(${STANDARD_TARGETS[@]})
elif [[ "$1" == "self" ]]; then
	REQUESTED_TARGETS=($GOOS/$GOARCH)
fi

if ! command -v godep 2>&1>/dev/null; then
	if ! go get github.com/tools/godep; then
		die "Unable to install missing dependency godep"
	fi
fi

log "Building for ${REQUESTED_TARGETS[@]}"

# Try to get the SHA that's being built from TeamCity first...
SHA="$BUILD_VCS_NUMBER"
# If that didn't work, sniff out the local git repo SHA
if [ -z "$SHA" ]; then
	if ! SHA=$(git rev-parse HEAD); then
		SHA="unknown"
	fi
	# Mark the build as dirty if any indexed files are modified
	if ! git diff-index --quiet HEAD; then
		SHA="dirty-$SHA"
	fi
fi

if [ -z "$BUILD_NUMBER" ]; then
	BUILD_NUMBER="unknown"
fi
# Try to get version number from branch name (TeamCity checks out tags as branches)
VERSION="HEAD"
if [ -z "$BRANCH" ]; then
	BRANCH="$(git branch | grep '^\*' | cut -d' ' -f2)"
fi
# Look for a branch name starting vN.N.N
#if (echo "$BRANCH" | grep '^v\d\+\.\d\+\.\d\+'); then
#	VERSION="$BRANCH"
#fi
TIMESTAMP="$(date +%s)"

BUILDS_FAILED=0
for T in ${REQUESTED_TARGETS[@]}; do
	IFS='/' read -ra PARTS <<< "$T"
	export GOOS="${PARTS[0]}" GOARCH="${PARTS[1]}"
	flags="-X main.CommitSHA=$SHA -X main.BuildNumber=$BUILD_NUMBER \
		-X main.Version=$VERSION \
		-X main.Branch=$BRANCH \
		-X main.BuildTimestamp=$TIMESTAMP \
		-X main.OS=$GOOS -X main.Arch=$GOARCH"
	ART_BASEDIR="$(pwd)/artifacts"
	ART_PATH="$ART_BASEDIR/$VERSION/$GOOS/$GOARCH"
	
	if ! godep go build -ldflags="$flags" -o "$ART_PATH/sous"; then
		log "Build failed for $T"
		((BUILDS_FAILED++))
		continue
	fi
	ARCHIVE_PATH="$ART_BASEDIR/sous-$VERSION-$GOOS-$GOARCH.tar.gz"
	# Create the archive
	if ! (cd $ART_PATH && tar czf "$ARCHIVE_PATH" sous); then
		log "Failed to create archive for $V"
		((BUILDS_FAILED++))
	fi
done

if [[ "$BUILDS_FAILED" == 1 ]]; then
	die "1 build failed."
elif [[ "$BUILDS_FAILED" != 0 ]]; then
	die "$BUILDS_FAILED builds failed"
fi

log "All done successfully."
