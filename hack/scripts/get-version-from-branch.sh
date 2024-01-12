#!/usr/bin/env bash

# This script will generate the version.
#
# First it verifies, that the current branch name is 'release-x.y',
# where x and y are multi-digit integers.
# It further looks into the existing tags, looking for ones that start with x.y.
# If there is none, it will return x.y.0. Otherwise it will return x.y.z where z
# is the highest existing value increase by one.
#
# Examples:
# 1. The current git branch is 'release-1.0' and git has the tags '1.0.1', '1.0.2', 1.0.3':
#
#    $ ./get-version-from-branch.sh
#    1.0.4
#
# 2. The current branch is 'release-1.1' and git has no tags that start with '1.1':
#
#    $ ./get-version-from-branch.sh
#    1.0.0
#
# 3. The current branch is 'main':
#    $ ./get-version-from-branch.sh
#    Not on a release branch.
#
# Please note that this will exit with an error:
#
#    ./get-version-from-branch.sh || echo "exit with error"
#    Not on a release branch.
#    exit with error

# Get the current branch name.
current_branch=$(git rev-parse --abbrev-ref HEAD)

# Check if the current branch is a release branch.
if [[ $current_branch =~ ^release-([0-9]+)\.([0-9]+)$ ]]; then
	# Extract x and y from the branch name. BASH_REMATCH is an array variable
	# automatically generated by pattern matching ([[ ... ]]).
	x=${BASH_REMATCH[1]}
	y=${BASH_REMATCH[2]}

	# Find the highest z value for the matching tags.
	highest_z=$(git tag -l "$x.$y.*" | cut -d '.' -f 3 | sort -n | tail -n 1)

	# Increment the highest z value by 1 or set to 0 if no matching tags are found.
	if [ -z "$highest_z" ]; then
		next_z=0
	else
		next_z=$((highest_z + 1))
	fi

	# Return the new version.
	new_version="${x}.${y}.${next_z}"
	echo "${new_version}"
	export NEW_VERSION=$new_version
	exit 0
else
	echo "Not on a release branch."
	exit 1
fi6
