#!/bin/bash
#
# Generate self-signed TLS certificates for local development
# Usage: ./scripts/generate-certs.sh [domain]
#

set -e

DOMAIN="${1:-localhost}"
CERT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../certs" && pwd)"

echo "Generating TLS certificates for domain: $DOMAIN"
echo "Certificate directory: $CERT_DIR"

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key and self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$CERT_DIR/server.key" \
    -out "$CERT_DIR/server.crt" \
    -subj "/C=VN/ST=Ho Chi Minh/L=HCM/O=Vietnam Stock Analysis/CN=$DOMAIN" \
    -addext "subjectAltName=DNS:$DOMAIN,DNS:localhost,IP:127.0.0.1"

# Set proper permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"

echo ""
echo "Certificates generated successfully!"
echo "  Certificate: $CERT_DIR/server.crt"
echo "  Private key: $CERT_DIR/server.key"
echo ""
echo "To use with curl (skip certificate verification):"
echo "  curl -k https://$DOMAIN:8080/health"
echo ""
echo "To install certificate system-wide (optional):"
echo "  # macOS"
echo "  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_DIR/server.crt"
echo "  # Linux (Debian/Ubuntu)"
echo "  sudo cp $CERT_DIR/server.crt /usr/local/share/ca-certificates/$DOMAIN.crt"
echo "  sudo update-ca-certificates"
