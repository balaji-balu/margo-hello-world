#!/bin/bash

MODE=${1:-process}

if [ "$MODE" == "process" ]; then
  ./scripts/test-process.sh
elif [ "$MODE" == "docker" ]; then
  ./scripts/test-docker.sh
else
  echo "Usage: $0 [process|docker]"
  exit 1
fi


