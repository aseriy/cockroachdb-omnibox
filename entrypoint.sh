#!/bin/bash

# Exit immediately if a command fails
set -e

# Setup a trap to kill all background processes if the script receives a SIGTERM/SIGINT
trap 'kill $(jobs -p) 2>/dev/null' EXIT

# Start process 1 in the background
./process_one &

# Start process 2 in the background
./process_two &

# Wait for ANY background process to exit or fail
wait -n

# Once wait -n unblocks, it means one process folded. 
# Exit the wrapper script, which triggers the trap above and kills the survivor.
exit 1
