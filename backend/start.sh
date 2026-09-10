#!/bin/sh
set -eu

config_path="${CMS_CONFIG_PATH:-./configs/config.yaml}"

echo "Starting database migration..."
./migrate -f "$config_path"

echo "Starting backend server..."
exec ./main -f "$config_path"
