#!/bin/sh
set -eu
exec ./main -f "${CMS_CONFIG_PATH:-./configs/config.yaml}"
