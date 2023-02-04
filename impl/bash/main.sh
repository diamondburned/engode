#!/usr/bin/env bash

main() {
	data=$(compress)
	dict=()

	while -n1 char; do

	done <<< "$data"
}

compress() { gzip -9c -; }
decompress() { gunzip -c -; }

main "$@"
