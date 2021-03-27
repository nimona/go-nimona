#!/bin/sh

# go/clangwrap.sh

SDK_PATH=`xcrun --sdk $SDK --show-sdk-path`
CLANG=`xcrun --sdk $SDK --find clang`

if [ "$GOARCH" == "amd64" ]; then
    CARCH="x86_64"
elif [ "$GOARCH" == "arm64" ]; then
    CARCH="arm64"
elif [ "$GOARCH" == "arm" ]; then
    CARCH="armv7"
fi

exec $CLANG -arch $CARCH -isysroot $SDK_PATH -mios-version-min=10.0 "$@"