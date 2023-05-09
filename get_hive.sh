#!/bin/bash
COMMIT=$1

if [ -z "$COMMIT" ]
then
    echo "ERROR: no commit hash given!"
    exit 1
fi

SUBMODULES=$(find . -name "go.mod" -printf '%h\n' | sed -e 's/^\.\///' | sort)

for submodule in $SUBMODULES
do
    cd "$submodule"
    echo "updating ${submodule}..."
    hivemodules=$(grep '^\sgithub.com/iotaledger/hive.go' go.mod | awk '{print $1}')
    for hivemodule in $hivemodules; do
        echo "   go get -u ${hivemodule}..."
        go get -u "$hivemodule@$COMMIT"
    done
    go mod tidy
    
    cd ..
done
