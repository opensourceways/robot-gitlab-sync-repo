#!/bin/bash

set -eu
# don't set pipefail, otherwise it will fail when find lfs files
# set -euo pipefail

work_dir=$1
repo_url=$2
repo_name=$3
commit=$4

echo_message() {
    echo "$1, $2, $3, $4"
}

cd $work_dir

git clone -q $repo_url
cd $repo_name

all_files=$work_dir/${commit}_files
git show --pretty="" --name-only $commit > $all_files
if [ ! -s $all_files ]; then
    echo_message "small" "no" "lfs" "no"

    exit 0
fi

lfs_files=$work_dir/${commit}_lfs
cat $all_files | xargs egrep "^oid sha256:.{64}$" > $lfs_files
if [ ! -s $lfs_files ]; then
    echo_message "$all_files" "yes" "lfs" "no"

    exit 0
fi

n=$(wc -l $all_files | awk '{print $1}')
if [ $n -eq 1 ]; then
    # When there is only one file, egrep will only output the string not the file:string
    # The content of $lfs_files is not file name, but the "oid sha256:xxx" on this case.
    echo "$(head -1 $all_files):$(head -1 $lfs_files)" > $lfs_files
    echo_message "small" "no" "$lfs_files" "yes"

    exit 0
fi

small_files=$work_dir/${commit}_small
cat $all_files $lfs_files | awk -F ':oid sha256:' '{print $1}' | sort | uniq -u > $small_files

v="no"
test -s $small_files && v="yes"

echo_message "$small_files" "$v" "$lfs_files" "yes"
