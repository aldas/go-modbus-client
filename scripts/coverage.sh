#!/usr/bin/env bash
#
# Code coverage generation

COVERAGE_DIR="${COVERAGE_DIR:-coverage}"
PKG_LIST=$(go list ./...)

# Create the coverage files directory
mkdir -p "$COVERAGE_DIR";

# Create a coverage file for each package
for package in ${PKG_LIST}; do
    go test -covermode=count -coverprofile "${COVERAGE_DIR}/${package##*/}.cov" "$package" ;
done ;

# Merge the coverage profile files
echo 'mode: count' > "${COVERAGE_DIR}"/coverage.xcov ;
tail -q -n +2 "${COVERAGE_DIR}"/*.cov >> "${COVERAGE_DIR}"/coverage.xcov ;

# Display the global code coverage
go tool cover -func="${COVERAGE_DIR}"/coverage.xcov ;

# If needed, generate HTML report
if [[ "$1" == "html" ]]; then
    go tool cover -html="${COVERAGE_DIR}"/coverage.xcov -o coverage.html ;
fi

# Remove the coverage files directory
rm -rf "$COVERAGE_DIR";
