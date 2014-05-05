#!/bin/sh

# OpenStack Env Variables
export OS_USERNAME="jhop"
export OS_API_KEY="a4f8a6f950c208fab35b128c43ae37e0"
export OS_AUTH_URL="https://identity.api.rackspacecloud.com/v2.0/tokens"
export OS_REGION_NAME="DFW"

# Docker specific settings
export DOCKER_CLI_PLUGINS="rax"

# Use the new binary to do our bidding with debug enabled
./bundles/0.10.0-dev/binary/docker -D $@;
