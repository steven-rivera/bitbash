#!/usr/bin/env bash

run_tests=false

while getopts "t" opt; do
	case "$opt" in
	t) run_tests=true ;;
	*)
		echo "Usage: $0 [-t]"
		exit 1
		;;
	esac
done

if $run_tests; then
	go test -v ./bitbash
else
	HISTFILE="history.txt" go run ./bitbash
fi