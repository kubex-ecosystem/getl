#!/usr/bin/env bash

set -e

_THIS_SCRIPT_PATH=$(realpath "$0")
_THIS_SCRIPT_DIR=$(dirname "${_THIS_SCRIPT_PATH}")

_EXAMPLE_FILE_NAME="exp_config_a.json"
_EXAMPLE_FILE_PATH="${_THIS_SCRIPT_DIR}/configFiles/${_EXAMPLE_FILE_NAME}"

_EXAMPLE_CMD="data-table"

_GETL_PATH="./getl"
_SUCCESS="\033[0;32m"
_WARN="\033[0;33m"
_ERROR="\033[0;31m"
_INFO="\033[0;36m"
_NC="\033[0m"

# Log messages with different levels
# Arguments:
#   $1 - log level (info, warn, error, success)
#   $2 - message to log
log() {
  local type=
  type=${1:-info}
  local message=
  message=${2:-}

  # With colors
  case $type in
    info|_INFO|-i|-I)
      printf '%b[_INFO]%b ℹ️  %s\n' "$_INFO" "$_NC" "$message"
      ;;
    warn|_WARN|-w|-W)
      printf '%b[_WARN]%b ⚠️  %s\n' "$_WARN" "$_NC" "$message"
      ;;
    error|_ERROR|-e|-E)
      printf '%b[_ERROR]%b ❌  %s\n' "$_ERROR" "$_NC" "$message"
      ;;
    success|_SUCCESS|-s|-S)
      printf '%b[_SUCCESS]%b ✅  %s\n' "$_SUCCESS" "$_NC" "$message"
      ;;
    *)
      log "info" "$message"
      ;;
  esac
}

# Check if the example file exists
check_example_file() {
  # Check if the script is run from the correct directory
  if [ ! -f "${_EXAMPLE_FILE_PATH}" ]; then
    return 1
  fi
  # Check if the example file exists
  log "info" "This script will use the example file: ${_EXAMPLE_FILE_PATH}"
  return 0
}

# Check if the script is run from the correct directory
# shellcheck disable=SC2016
get_getl() {
  if [ ! -f "${_GETL_PATH}" ]; then
    if ! command -v getl &> /dev/null; then
      log "info" "getl could not be found, please install it first."
      log "info" "You can install it by running the following command:"
      log "info" 'curl -sSL "https://getl.dev/install.sh" | bash install && . "$HOME/.$(basename ${SHELL})rc"'
      log "info" "The script will do all the work for you. After installation, you can run this script again."
      log "info" "See you later! 👋"
      return 1
    else
      _GETL_PATH=$(command -v getl)
    fi
  fi
  # Check if getl is executable
  if [ ! -x "${_GETL_PATH}" ]; then
    log "error" "getl is not executable. Please check your installation."
    return 1
  fi
  _GETL_PATH=$(realpath "${_GETL_PATH}")
  return 0
}

# Validate the example file and getl installation
validate_files() {
  # Check if the script is run from the correct directory
  if ! check_example_file; then
    log "error" "Example file not found. Please run this script from the root directory of the project."
    log "error" "You are currently in: $(pwd)"
    log "error" "Please run the script from the root directory of the project."
    exit 1
  fi

  # Check if getl is installed
  # Log is inside the function
  get_getl || exit 1
}

case $1 in
  --help|-h)
    log "info" "Usage: bash ${_THIS_SCRIPT_PATH}"
    log "info" "Make sure to run from the project's root directory."
    exit 0
    ;;
esac

# Check if the script is run from the correct directory
validate_files || exit 1

# Construct the command
_cmd_to_run=("${_GETL_PATH}" "${_EXAMPLE_CMD}" "-f" "${_EXAMPLE_FILE_PATH}")

# Run the command
log "info" "Running command: ${_cmd_to_run[*]}"

# Execute the command
"${_cmd_to_run[@]}"

# Exit with the status of the last command
exit $?