#!/usr/bin/env bash
# Source: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

platforms=(
    "windows/amd64" "windows/386" "linux/amd64" "linux/386"
    "linux/arm" "linux/arm64" "freebsd/amd64" "freebsd/386"
    "netbsd/amd64" "netbsd/386" "openbsd/amd64" "openbsd/386"
    "plan9/amd64" "plan9/386" "darwin/amd64" "darwin/arm64"
)

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name='mockserver-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi	

	env GOOS=$GOOS GOARCH=$GOARCH go build -o ./bin/$output_name ./cmd/mockserver
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done
