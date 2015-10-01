#!/usr/bin/env bash

if SOUSPATH="$(command -v sous)"; then
	if [[ "$SOUSPATH" == "$GOPATH/bin/sous" ]]; then
		rm "$SOUSPATH"
	else
		if (command -v brew 2>&1>/dev/null); then
			if (brew list | grep '\bsous\b'); then
				echo "Detected sous installed by homebrew, uninstalling..."
				if ! brew uninstall sous; then
					echo "Uninstall failed."
					exit 1
				fi
			fi
		else
			echo "Existing sous not recognised, please remove $SOUSPATH"
			exit 1
		fi
	fi
fi
VERSION="$(id -un)@$(hostname)-$(date '+%Y-%m-%dT%H:%M:%S')"
FLAGS="-X main.OS=$GOOS -X main.Arch=$GOARCH -X main.Version=$VERSION"
FLAGS="$FLAGS  -X main.CommitSHA=$(git rev-parse HEAD)"
if ! go install -ldflags "$FLAGS"; then
	echo "Build failed."
	exit 1
else
	echo "Done!"
	sous version
fi
