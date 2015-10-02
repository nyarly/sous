#!/usr/bin/env bash
set -euo pipefail
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
set +u
SHA="$BUILD_VCS_NUMBER"
set -u
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
set +u
if [ -z "$BUILD_NUMBER" ]; then
	BUILD_NUMBER="unknown"
fi
# Try to get version number from branch name (TeamCity checks out tags as branches)
if [ -z "$BRANCH" ]; then
	BRANCH="$(git branch | grep '^\*' | cut -d' ' -f2)"
fi
set -u
# Look for a branch name starting vN.N.N
#if (echo "$BRANCH" | grep '^v\d\+\.\d\+\.\d\+'); then
#	VERSION="$BRANCH"
#fi
VERSION="$(basename $BRANCH)"
TIMESTAMP="$(date +%s)"

log "Building sous version $VERSION; branch: $BRANCH; build number: $BUILD_NUMBER; Revision: $SHA"

# Empty the artifacts dir...
if [ -d ./artifacts ]; then
	rm -rf ./artifacts
	mkdir ./artifacts
	echo "Do not check in this directory, it is used for ephemeral build artifacts." > ./artifacts/README.md
fi

log "Cleaned artifacts directory."

BUILDS_FAILED=0
BUILDS_SUCCEEDED=0
for T in ${REQUESTED_TARGETS[@]}; do
	log "Starting compile for $T"
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
	log "Compile successful."
	ARCHIVE_PATH="$ART_BASEDIR/sous-$VERSION-$GOOS-$GOARCH.tar.gz"
	# Create the archive
	log "Archiving $ART_PATH as $ARCHIVE_PATH"
	if ! [ -d "$ART_PATH" ]; then
		log "Archive path does not exist: $ART_PATH"
		((BUILDS_FAILED++))
		continue
	fi
	if ! (cd $ART_PATH && tar -czvf "$ARCHIVE_PATH" .); then
		log "Failed to create archive for $V"
		((BUILDS_FAILED++))
		continue
	fi
	# Write homebrew bottles
	if [[ "$GOOS" == "darwin" ]]; then
		log "Detected darwin (Mac OS X) build; generating Homebrew bottles..."
		for OSX_VERSION in el_capitan yosemite mavericks mountain_lion; do
			BOTTLE_PATH="$ART_BASEDIR/sous${VERSION%v}.${OSX_VERSION}.bottle.1.tar.gz"
			log "Bottling $VERSION for $OSX_VERSION..."
			cp "$ARCHIVE_PATH" "$BOTTLE_PATH"
			openssl dgst -sha256 "$BOTTLE_PATH"
		done
		log "Bottles built, see digests above."
	fi
	((BUILDS_SUCCEEDED++))
done
TOTAL_BUILDS=$((BUILDS_SUCCEEDED+BUILDS_FAILED))
if [[ "$BUILDS_FAILED" == 1 ]]; then
	die "1 build of $TOTAL_BUILDS failed."
elif [[ "$BUILDS_FAILED" != 0 ]]; then
	die "$BUILDS_FAILED of $TOTAL_BUILDS builds failed"
fi

log "========================= Contents of $ART_BASEDIR:"
ls -lah "$ART_BASEDIR"
log "========================= END"


log "All $BUILDS_SUCCEEDED of $BUILDS_SUCCEEDED builds were successful."
