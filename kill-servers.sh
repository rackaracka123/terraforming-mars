#!/bin/bash

# Terraforming Mars - Kill Development Servers Script
# Terminates all frontend and backend processes for the project

set -e

PROJECT_DIR="/home/mafs/Documents/Repositories/terraforming-mars"
echo "🛑 Terminating Terraforming Mars development servers..."

# Function to kill processes by name pattern with error handling
kill_by_pattern() {
    local pattern="$1"
    local description="$2"

    local pids=$(pgrep -f "$pattern" 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔪 Killing $description processes: $pids"
        echo "$pids" | xargs kill -TERM 2>/dev/null || true
        sleep 1
        # Force kill if still running
        local remaining=$(pgrep -f "$pattern" 2>/dev/null || true)
        if [ -n "$remaining" ]; then
            echo "💥 Force killing remaining $description processes: $remaining"
            echo "$remaining" | xargs kill -KILL 2>/dev/null || true
        fi
    else
        echo "ℹ️  No $description processes found"
    fi
}

# Function to kill processes using specific ports
kill_by_port() {
    local port="$1"
    local description="$2"

    local pids=$(lsof -t -i:$port 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔪 Killing processes using port $port ($description): $pids"
        echo "$pids" | xargs kill -TERM 2>/dev/null || true
        sleep 1
        # Force kill if still running
        local remaining=$(lsof -t -i:$port 2>/dev/null || true)
        if [ -n "$remaining" ]; then
            echo "💥 Force killing remaining processes on port $port: $remaining"
            echo "$remaining" | xargs kill -KILL 2>/dev/null || true
        fi
    else
        echo "ℹ️  No processes using port $port"
    fi
}

# Kill frontend processes (npm, vite, esbuild)
kill_by_pattern "npm start.*terraforming-mars" "npm start"
kill_by_pattern "node.*vite" "vite dev server"
kill_by_pattern "esbuild.*terraforming-mars" "esbuild"

# Kill any vite process in this project directory (more specific)
cd "$PROJECT_DIR" && {
    local_vite_pids=$(pgrep -f "node.*vite" 2>/dev/null | while read pid; do
        if [ -n "$(lsof -p $pid 2>/dev/null | grep $(pwd))" ]; then
            echo $pid
        fi
    done)

    if [ -n "$local_vite_pids" ]; then
        echo "🔪 Killing project-specific vite processes: $local_vite_pids"
        echo "$local_vite_pids" | xargs kill -TERM 2>/dev/null || true
        sleep 1
        # Check if any are still running and force kill
        for pid in $local_vite_pids; do
            if kill -0 $pid 2>/dev/null; then
                echo "💥 Force killing remaining vite process: $pid"
                kill -KILL $pid 2>/dev/null || true
            fi
        done
    fi
}

# Kill backend Go processes
kill_by_pattern "go run.*cmd/server/main.go" "Go backend server"
kill_by_pattern "go run.*cmd/watch" "Go watch server"

# Kill processes by ports (fallback)
kill_by_port 3000 "frontend"
kill_by_port 3001 "backend"

# Additional cleanup for any lingering processes in the project directory
cd "$PROJECT_DIR" 2>/dev/null || true
local_pids=$(pgrep -f "$PROJECT_DIR" 2>/dev/null | grep -E "(npm|node|go)" || true)
if [ -n "$local_pids" ]; then
    echo "🧹 Cleaning up remaining project processes: $local_pids"
    echo "$local_pids" | xargs kill -TERM 2>/dev/null || true
    sleep 1
    local_remaining=$(pgrep -f "$PROJECT_DIR" 2>/dev/null | grep -E "(npm|node|go)" || true)
    if [ -n "$local_remaining" ]; then
        echo "💥 Force killing remaining project processes: $local_remaining"
        echo "$local_remaining" | xargs kill -KILL 2>/dev/null || true
    fi
fi

echo ""
echo "✅ Server termination complete!"
echo ""
echo "🔍 Final check - processes still using development ports:"
lsof -i :3000,3001 2>/dev/null || echo "   ✅ No processes using ports 3000 or 3001"

echo ""
echo "📋 To start servers again, run: make run"