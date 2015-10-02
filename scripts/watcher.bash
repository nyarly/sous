#!/usr/bin/env bash
s() { say -v Daniel "$@" & }
fin() { s "Bye!"; exit 0; }
build() {
	result=$(./scripts/install-dev.bash)
	s $(echo "$result" | tail -n1)
	echo "$result"
}
trap fin SIGINT SIGTERM
while true; do
	fswatch -ro . | build
done

