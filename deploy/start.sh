#!/bin/sh
set -e

# Start Go backend on internal port 3001
PORT=3001 ./server &
GO_PID=$!

# Wait briefly for the Go backend to start
sleep 1

# Start Next.js frontend on Dokku's assigned PORT (default 5000)
export API_URL="http://localhost:3001"
export HOSTNAME="0.0.0.0"

node server.js &
NODE_PID=$!

# Handle shutdown: forward signals to both processes
trap 'kill $GO_PID $NODE_PID 2>/dev/null; wait' SIGTERM SIGINT

# Wait for either process to exit
wait -n $GO_PID $NODE_PID 2>/dev/null || true

# If one exits, kill the other
kill $GO_PID $NODE_PID 2>/dev/null
wait
