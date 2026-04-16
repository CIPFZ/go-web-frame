#!/bin/sh

echo "Waiting for database to start..."
sleep 15

echo "Starting database migration..."
./migrate -f ./configs/config.yaml

echo "Starting backend server..."
./main -f ./configs/config.yaml
