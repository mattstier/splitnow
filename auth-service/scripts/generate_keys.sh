#!/usr/bin/env bash
set -euo pipefail

KEYS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/keys"
mkdir -p "$KEYS_DIR"

openssl genpkey -algorithm RSA \
	-out "$KEYS_DIR/jwt_private.pem" \
	-pkeyopt rsa_keygen_bits:2048

openssl rsa \
	-in "$KEYS_DIR/jwt_private.pem" \
	-pubout \
	-out "$KEYS_DIR/jwt_public.pem"

echo "keys written to $KEYS_DIR"
