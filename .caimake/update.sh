#!/bin/bash

# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in pipeline.
set -o pipefail

# Controls verbosity of the script output and logging.
VERBOSE="${VERBOSE:-5}"

# If set true, the log will be colorized
readonly COLOR_LOG=${COLOR_LOG:-true}

readonly CAIMAKE_HOME=${CAIMAKE_HOME:-${HOME}/.caimake}

readonly CAIMAKE_AUTO_UPDATE=${CAIMAKE_AUTO_UPDATE:-false}

readonly project="caicloud/build-infra"

if [[ ${COLOR_LOG} == "true" ]]; then
	readonly blue="\033[34m"
	readonly green="\033[32m"
	readonly red="\033[31m"
	readonly yellow="\033[36m"
	readonly strong="\033[1m"
	readonly reset="\033[0m"
else
	readonly blue=""
	readonly green=""
	readonly red=""
	readonly yellow=""
	readonly strong=""
	readonly reset=""
fi

# Print a status line.  Formatted to show up in a stream of output.
log::status() {
	local V="${V:-0}"
	if [[ $VERBOSE < $V ]]; then
		return
	fi

	timestamp=$(date +"[%m%d %H:%M:%S]")
	echo -e "${blue}==> $timestamp${reset} ${strong}$1${reset}"
	shift
	for message; do
		echo "    $message"
	done
}

# Log an error but keep going.  Don't dump the stack or exit.
log::error() {
	timestamp=$(date +"[%m%d %H:%M:%S]")
	echo -e "${red}!!! $timestamp${reset} ${strong}${1-}${reset}"
	shift
	for message; do
		echo "    $message"
	done
}

command_exists() {
	command -v "$@" >/dev/null 2>&1
}

caimake::update() {
	mkdir -p ${CAIMAKE_HOME}

	if ! command_exists caimake; then
		# add CAIMAKE_HOME to PATH
		export PATH=${CAIMAKE_HOME}:${PATH}
		if ! command_exists caimake; then
			# check again
			caimake::download_binary
		fi
	fi

	if [[ ${CAIMAKE_AUTO_UPDATE-} == "true" ]]; then
		caimake update
	fi
}

caimake::get_version_from_url() {
	local url=${1-}
	local version=${url##*/releases/download/}
	version=${version%/*}
	echo ${version}
}

caimake::download_binary() {
	local hostos="$(uname -s | tr '[A-Z]' '[a-z]')"
	local latest="$(curl -s https://api.github.com/repos/${project}/releases/latest | grep browser_download_url | cut -d '"' -f 4 | grep -m 1 ${hostos})"
	local version="$(caimake::get_version_from_url ${latest-})"

	if [[ -z ${version} ]]; then
		# no release found on github
		log::error "No caimake release found on Github"
		exit 1
	fi

	log::status "Latest caimake release on Github is ${version}"

	curl -L ${latest} -o ${CAIMAKE_HOME}/caimake.temp
	mv ${CAIMAKE_HOME}/caimake.temp ${CAIMAKE_HOME}/caimake
	chmod +x ${CAIMAKE_HOME}/caimake
}
