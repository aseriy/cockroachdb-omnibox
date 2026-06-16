#!/bin/bash

# Exit immediately if a command fails
set -e

# Setup a trap to kill all background processes if the script receives a SIGTERM/SIGINT
trap 'kill $(jobs -p) 2>/dev/null' EXIT

JOIN="tasks.${SELF}:26258"
for n in ${NEIGHBORS}; do
    JOIN="${JOIN},tasks.${n}:26258"
done

# Start process 1 in the background
/cockroach/cockroach start \
    --certs-dir=/cockroach/certs \
    --listen-addr=:26258 --sql-addr=:26257 --http-addr=:8080 \
    --advertise-addr=tasks.${SELF}:26258 \
    --advertise-sql-addr=tasks.${SELF}:26257 \
    --join=${JOIN} \
    --store=/data &

# # Start process 2 in the background
# dbworkload run -w DatapointLogger.py --uri  &

# Wait for ANY background process to exit or fail
wait -n

# Once wait -n unblocks, it means one process folded. 
# Exit the wrapper script, which triggers the trap above and kills the survivor.
exit 1
