#!/bin/sh

echo "Waiting for MySQL to start..."
sleep 15

echo "Starting database migration..."
./migrate -f ./configs/config.yaml

echo "Starting backend server..."
./main -f ./configs/config.yaml
