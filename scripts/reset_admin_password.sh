#!/bin/bash

# OpenWan Admin Password Reset Script
# Usage: ./reset_admin_password.sh [new_password]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database configuration
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_USER="${DB_USER:-openwan}"
DB_PASS="${DB_PASS:-openwan123}"
DB_NAME="${DB_NAME:-openwan_db}"

# Default password
DEFAULT_PASSWORD="admin123"
NEW_PASSWORD="${1:-$DEFAULT_PASSWORD}"

echo -e "${YELLOW}==================================${NC}"
echo -e "${YELLOW}OpenWan Admin Password Reset Tool${NC}"
echo -e "${YELLOW}==================================${NC}"
echo ""

# Validate password length
if [ ${#NEW_PASSWORD} -lt 6 ]; then
    echo -e "${RED}Error: Password must be at least 6 characters${NC}"
    exit 1
fi

if [ ${#NEW_PASSWORD} -gt 32 ]; then
    echo -e "${RED}Error: Password must be at most 32 characters${NC}"
    exit 1
fi

echo -e "Database: ${GREEN}$DB_HOST/$DB_NAME${NC}"
echo -e "User: ${GREEN}admin${NC}"
echo -e "New password: ${GREEN}$NEW_PASSWORD${NC}"
echo ""

# Check database connectivity
echo -n "Checking database connectivity... "
if mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME" 2>/dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}Cannot connect to database. Check credentials.${NC}"
    exit 1
fi

# Generate bcrypt hash using Go
echo -n "Generating password hash... "
cd "$(dirname "$0")"

# Create temporary Go script
cat > /tmp/hash_password.go << 'GOEOF'
package main

import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: hash_password <password>")
        os.Exit(1)
    }
    
    password := os.Args[1]
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Print(string(hash))
}
GOEOF

# Generate hash
PASSWORD_HASH=$(go run /tmp/hash_password.go "$NEW_PASSWORD" 2>/dev/null)
if [ $? -ne 0 ]; then
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}Cannot generate password hash${NC}"
    rm -f /tmp/hash_password.go
    exit 1
fi
rm -f /tmp/hash_password.go

echo -e "${GREEN}OK${NC}"
echo "Hash: ${PASSWORD_HASH:0:30}..."

# Update database
echo -n "Updating admin password... "
mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "
UPDATE ow_users 
SET password = '$PASSWORD_HASH', enabled = 1
WHERE username = 'admin';
" 2>/dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    exit 1
fi

# Verify update
echo -n "Verifying update... "
RESULT=$(mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" -N -e "
SELECT COUNT(*) FROM ow_users 
WHERE username = 'admin' AND enabled = 1;
" 2>/dev/null)

if [ "$RESULT" = "1" ]; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}==================================${NC}"
echo -e "${GREEN}Password reset successful!${NC}"
echo -e "${GREEN}==================================${NC}"
echo ""
echo "Credentials:"
echo "  Username: admin"
echo "  Password: $NEW_PASSWORD"
echo "  Status: Enabled"
echo ""
echo "You can now login at: http://localhost:3000"
echo ""

# Test login (optional)
read -p "Test login with new password? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -n "Testing login... "
    RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"admin\",\"password\":\"$NEW_PASSWORD\"}")
    
    SUCCESS=$(echo "$RESPONSE" | grep -o '"success":[^,}]*' | cut -d':' -f2)
    
    if [ "$SUCCESS" = "true" ]; then
        echo -e "${GREEN}SUCCESS${NC}"
        echo "Login test passed!"
    else
        echo -e "${RED}FAILED${NC}"
        echo "Response: $RESPONSE"
        echo ""
        echo -e "${YELLOW}Note: Password was updated in database, but login test failed.${NC}"
        echo -e "${YELLOW}This might be due to backend service not running.${NC}"
    fi
fi

echo ""
echo "Done!"
