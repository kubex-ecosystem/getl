#!/usr/bin/env bash

# set -o posix
set -o nounset   # Treat unset variables as an error
set -o errexit   # Exit immediately if a command exits with a non-zero status
set -o pipefail  # Prevent errors in a pipeline from being masked
set -o errtrace  # If a command fails, the shell will exit immediately
set -o functrace # If a function fails, the shell will exit immediately
# shopt -s inherit_errexit # Inherit the errexit option in functions
IFS=$'\n\t'

_ROOT_DIR="${_ROOT_DIR:-}"
_APP_NAME="${_APP_NAME:-}"
_DESCRIPTION="${_DESCRIPTION:-}"
_OWNER="${_OWNER:-}"
_BINARY_NAME="${_BINARY_NAME:-}"
_PROJECT_NAME="${_PROJECT_NAME:-}"
_AUTHOR="${_AUTHOR:-}"
_VERSION="${_VERSION:-}"
_LICENSE="${_LICENSE:-}"
_REPOSITORY="${_REPOSITORY:-}"
_PRIVATE_REPOSITORY="${_PRIVATE_REPOSITORY:-}"
_VERSION_GO="${_VERSION_GO:-}"
_PLATFORMS_SUPPORTED="${_PLATFORMS_SUPPORTED:-}"

_MANIFEST_SUBPATH=${_MANIFEST_SUBPATH:-'internal/module/info/manifest.json'}

__get_values_from_manifest() {
  # # Define the root directory (assuming this script is in lib/ under the root)
  _ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

  # shellcheck disable=SC2005
  _APP_NAME="$(jq -r '.bin' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "$(basename "${_ROOT_DIR:-}")")"
  _DESCRIPTION="$(jq -r '.description' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "No description provided.")"
  _OWNER="$(jq -r '.organization' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "kubex-ecosystem")"
  _OWNER="${_OWNER,,}" # Converts to lowercase
  _BINARY_NAME="${_APP_NAME}"
  _PROJECT_NAME="$(jq -r '.name' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "$_APP_NAME")"
  _AUTHOR="$(jq -r '.author' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "Rafa Mori")"
  _VERSION=$(jq -r '.version' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "v0.0.0")
  _LICENSE="$(jq -r '.license' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "MIT")"
  _REPOSITORY="$(jq -r '.repository' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "kubex-ecosystem/${_APP_NAME}")"
  _PRIVATE_REPOSITORY="$(jq -r '.private' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "false")"
  _VERSION_GO="$(jq -r '.go_version' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || grep '^go ' "$_ROOT_DIR/go.mod" | awk '{print $2}' || echo "")"
  _VERSION_GO="${_VERSION_GO//null/}" # Remove 'null' if present
  _PLATFORMS_SUPPORTED="$(jq -r '.platforms[]' "$_ROOT_DIR/$_MANIFEST_SUBPATH" 2>/dev/null || echo "linux, macOS, windows")"
  _PLATFORMS_SUPPORTED="$(printf '%s ' "${_PLATFORMS_SUPPORTED[*]//
/, }")"                                            # Converts to comma-separated list
  _PLATFORMS_SUPPORTED="${_PLATFORMS_SUPPORTED,,}" # Converts to lowercase

  return 0
}

apply_manifest() {
  __get_values_from_manifest || return 1
  return 0
}

export -f apply_manifest
