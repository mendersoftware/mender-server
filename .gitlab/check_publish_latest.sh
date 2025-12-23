#!/bin/sh

if test -z "$CI_COMMIT_TAG"; then
	exit 1
fi

git fetch --tags

LATEST_TAG=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' |
	grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' |
	sort -t. -k1,1n -k2,2n -k3,3n |
	tail -n1)

echo "Current tag: $CI_COMMIT_TAG"
echo "Latest tag:  $LATEST_TAG"

if test "$CI_COMMIT_TAG" = "$LATEST_TAG"; then
	exit 0
fi
exit 1
