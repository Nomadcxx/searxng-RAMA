#!/bin/bash
# Check for hardcoded secrets and sensitive information

echo "Checking for hardcoded secrets and sensitive information..."

# Secret patterns to check for
PATTERNS=(
    "secret_key:"
    "api_key"
    "password"
    "token"
    "secret"
    "key.*="
    "credentials"
    "private_key"
    "public_key"
    "ssh-rsa"
    "BEGIN.*PRIVATE"
    "access_token"
    "refresh_token"
    "client_secret"
    "aws_access_key"
    "aws_secret_key"
)

ERRORS=0

# Check configuration and code files
for pattern in "${PATTERNS[@]}"; do
    echo "Checking for pattern: $pattern"
    
    # Search in code files
    for file in $(find . -name "*.yml" -o -name "*.yaml" -o -name "*.py" -o -name "*.go" -o -name "*.sh" -o -name "*.json" -type f -not -path "./.git/*" -not -path "./node_modules/*" -not -path "./vendor/*" -not -path "./searxng-custom/*" 2>/dev/null); do
        # Skip known false positives
        if [[ "$file" == *"settings.yml"* ]] && [[ "$pattern" == "secret_key:" ]]; then
            # Check if it's the default placeholder
            if grep -q "secret_key: \"ultrasecretkey\"" "$file" 2>/dev/null; then
                continue
            fi
        fi
        
        if grep -q -i "$pattern" "$file" 2>/dev/null; then
            echo "  ⚠️  Found potential secret pattern \"$pattern\" in $file"
            grep -n -i "$pattern" "$file" 2>/dev/null | while read -r line; do
                echo "    $line"
                
                # Check if it's likely an actual hardcoded secret (not a comment or example)
                if echo "$line" | grep -q -v "#.*$pattern" && echo "$line" | grep -q ":\|=\| " && ! echo "$line" | grep -q "example\|test\|demo\|placeholder\|TODO\|FIXME"; then
                    echo "    ❌ This looks like a hardcoded secret (not a placeholder)!"
                    ERRORS=$((ERRORS + 1))
                fi
            done
        fi
    done
done

# Check for email addresses
echo "Checking for email addresses..."
find . -type f \( -name "*.py" -o -name "*.go" -o -name "*.sh" -o -name "*.md" -o -name "*.txt" \) -not -path "./.git/*" -not -path "./node_modules/*" -not -path "./vendor/*" 2>/dev/null | \
    xargs grep -E -l '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}' 2>/dev/null | \
    while read -r file; do
        echo "  ⚠️  Found email addresses in $file"
        grep -n -E '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}' "$file" 2>/dev/null
    done

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "❌ Found $ERRORS potential hardcoded secrets"
    echo "Please replace hardcoded secrets with environment variables or placeholders."
    exit 1
else
    echo "✅ No hardcoded secrets found (only acceptable placeholders)"
    exit 0
fi