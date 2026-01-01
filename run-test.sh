#!/bin/bash
set -e

# Clean up any existing test environment
rm -rf .epicenv

# 1. Create a test environment
./epicenv init danthegoodman1 -e local

# 2. Set some test variables
./epicenv set TEST_VAR "hello_world" -e local
./epicenv set ANOTHER_VAR "12345" -e local

# 3. Test run command with echo
echo "=== Testing run with echo ==="
OUTPUT=$(./epicenv -e local run -- bash -c 'echo $TEST_VAR')
if [ "$OUTPUT" = "hello_world" ]; then
    echo "✓ TEST_VAR correctly injected"
else
    echo "✗ TEST_VAR was '$OUTPUT', expected 'hello_world'"
    exit 1
fi

# 4. Test that multiple vars are injected
echo "=== Testing multiple vars ==="
OUTPUT=$(./epicenv -e local run -- bash -c 'echo $TEST_VAR-$ANOTHER_VAR')
if [ "$OUTPUT" = "hello_world-12345" ]; then
    echo "✓ Multiple vars correctly injected"
else
    echo "✗ Got '$OUTPUT', expected 'hello_world-12345'"
    exit 1
fi

# 5. Test run with a simple command (true)
echo "=== Testing exit code passthrough (success) ==="
./epicenv -e local run true
echo "✓ Exit code 0 passed through"

# 6. Test that exit codes are passed through
echo "=== Testing exit code passthrough (failure) ==="
if ./epicenv -e local run false; then
    echo "✗ Should have exited with non-zero"
    exit 1
else
    echo "✓ Non-zero exit code passed through"
fi

# 7. Test with overlay environment
echo "=== Testing with overlay environment ==="
./epicenv init -e staging --overlay local
./epicenv set TEST_VAR "staging_value" -e staging

OUTPUT=$(./epicenv -e staging run -- bash -c 'echo $TEST_VAR-$ANOTHER_VAR')
if [ "$OUTPUT" = "staging_value-12345" ]; then
    echo "✓ Overlay stacking works with run"
else
    echo "✗ Got '$OUTPUT', expected 'staging_value-12345'"
    exit 1
fi

# 8. Test that existing env vars are preserved
echo "=== Testing existing env var preservation ==="
export EXISTING_VAR="should_still_exist"
OUTPUT=$(./epicenv -e local run -- bash -c 'echo $EXISTING_VAR')
if [ "$OUTPUT" = "should_still_exist" ]; then
    echo "✓ Existing env vars preserved"
else
    echo "✗ Existing var was '$OUTPUT', expected 'should_still_exist'"
    exit 1
fi

echo ""
echo "=== All run command tests passed! ==="
