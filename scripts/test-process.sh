#!/bin/bash
set -e

echo "setup process script"
echo "🧩 Starting DEV Process Mode..."

# --- Directories ---
ROOT_DIR=$(dirname "$(realpath "$0")")/..
BIN_DIR="$ROOT_DIR/bin"
LOG_DIR="$ROOT_DIR/logs"
CONFIG_DIR="$ROOT_DIR/configs"
PIDS_FILE="$LOG_DIR/pids.txt"

mkdir -p "$BIN_DIR" "$LOG_DIR"

# --- Build all binaries ---
echo "⚙️  Building all binaries..."
MODE=${1:-debug}

if [ "$MODE" = "release" ]; then
    LDFLAGS="-ldflags=-s -w"
    echo "🚀 Building in RELEASE mode..."
else
    LDFLAGS=""
    echo "🔧 Building in DEBUG mode..."
fi

go build $LDFLAGS -o "$BIN_DIR/co" "$ROOT_DIR/cmd/co"
go build $LDFLAGS -o "$BIN_DIR/lo" "$ROOT_DIR/cmd/lo"
go build $LDFLAGS -o "$BIN_DIR/en" "$ROOT_DIR/cmd/en"


echo "✅ Build complete."

# --- Clear old PIDs ---
: > "$PIDS_FILE"

# --- Start processes ---
echo "🚀 Starting processes..."
"$BIN_DIR/co" --config="$CONFIG_DIR/co.yaml" > "$LOG_DIR/co.log" 2>&1 &
echo $! >> "$PIDS_FILE"

"$BIN_DIR/lo" --site site1 --eoport 8083 --config="$CONFIG_DIR/lo1.yaml" > "$LOG_DIR/lo1.log" 2>&1 &
echo $! >> "$PIDS_FILE"

"$BIN_DIR/lo" --site site2 --eoport 8082 --config="$CONFIG_DIR/lo2.yaml" > "$LOG_DIR/lo2.log" 2>&1 &
echo $! >> "$PIDS_FILE"

"$BIN_DIR/en" --node edge1 --port :8083 --config="$CONFIG_DIR/edge1.yaml" > "$LOG_DIR/edge1.log" 2>&1 &
echo $! >> "$PIDS_FILE"

"$BIN_DIR/en" --node edge2 --port :8082 --config="$CONFIG_DIR/edge2.yaml" > "$LOG_DIR/edge2.log" 2>&1 &
echo $! >> "$PIDS_FILE"

echo "✅ All processes started."
echo "Logs in $LOG_DIR"
echo "Press Ctrl+C to stop everything."

# --- Cleanup on Ctrl+C ---
cleanup() {
    echo ""
    echo "🛑 Stopping all processes..."
    while read -r PID; do
        if kill -0 "$PID" 2>/dev/null; then
            echo "Killing PID $PID"
            kill "$PID" 2>/dev/null || true
        fi
    done < "$PIDS_FILE"
    echo "✅ All processes stopped."
    exit 0
}
trap cleanup SIGINT SIGTERM

# --- Keep running ---
while true; do sleep 1; done
