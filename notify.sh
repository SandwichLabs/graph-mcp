#!/bin/bash

# Default topic will be generated if no topic is provided
TOPIC=""

usage() {
    echo "Usage: $0 [-t <topic>] <message>"
    exit 1
}

# Parse options
while getopts "t:" opt; do
  case $opt in
    t)
      TOPIC="$OPTARG"
      ;;
    \?)
      usage
      ;;
  esac
done

# Shift off the options
shift "$((OPTIND-1))"

# Check for message
if [ -z "$1" ]; then
  usage
fi

MESSAGE="$1"

# If topic is not provided, generate a random one
if [ -z "$TOPIC" ]; then
    TOPIC="jules-temp-topic-$(date +%s)-${RANDOM}"
fi

# Handle full URL topic
if [[ "$TOPIC" == https://* || "$TOPIC" == http://* ]]; then
    TOPIC=$(basename "$TOPIC")
fi

curl -d "$MESSAGE" "ntfy.sh/$TOPIC"

echo
echo "Notification sent to topic: $TOPIC"
echo "You can view it at https://ntfy.sh/$TOPIC"
