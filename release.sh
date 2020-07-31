#!/bin/bash

VERSION=$(head  -n 1 release.history)
VERSION=${VERSION#"- "}

VERSION_CODE="goopt.Version = \"${VERSION}\""
echo "VERSION     : ${VERSION}"
echo "VERSION_CODE: ${VERSION_CODE}"


sed -i "s~goopt\.Version = \".*\"~goopt.Version = \"${VERSION}\"~g" readme/main.go
sed -i "s~goopt\.Version = \".*\"~goopt.Version = \"${VERSION}\"~g" main.go
sed -i "s~goopt\.Version = \".*\"~goopt.Version = \"${VERSION}\"~g" _test/dbmeta/main.go

ack "goopt.Version = \".*\""
