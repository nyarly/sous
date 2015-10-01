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

log "Building for ${REQUESTED_TARGETS[@]}"


# Try to get the SHA that's being built
if ! SHA=$(git rev-parse HEAD); then
	SHA="unknown"
fi
if ! git diff-index --quiet HEAD; then
	SHA="dirty-$SHA"
fi

if [ -z "$BUILD_NUMBER" ]; then
	BUILD_NUMBER="unknown"
fi
# Try to get version number from branch name (TeamCity checks out tags as branches)
VERSION="unknown"
if BRANCH="$(git branch | grep '^\*' | cut -d' ' -f2)"; then
	# Look for a branch name starting vN.N.N
	if (echo $BRANCH | grep '^v\d\+\.\d\+\.\d\+'); then
		VERSION="$BRANCH"
	fi
fi
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
	if ! go build -ldflags="$flags" -o "artifacts/$T/sous"; then
		log "Build failed for $T"
		((BUILDS_FAILED++))
		continue
	fi
done
