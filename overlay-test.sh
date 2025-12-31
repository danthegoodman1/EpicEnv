#!/bin/bash
set -e

# Clean up any existing test environment
rm -rf .epicenv

# 1. Create base "local" environment
./epicenv init danthegoodman1 -e local

# 2. Create "testing" overlay on top of "local"
./epicenv init -e testing --overlay local

# 3. Create "agent-testing" overlay on top of "testing" (3-layer stack)
./epicenv init -e agent-testing --overlay testing

# 4. Set secrets at different layers
./epicenv set DB_HOST localhost -e local
./epicenv set S3_BUCKET s3://local-bucket -e local
./epicenv set LOG_LEVEL debug -e local

./epicenv set S3_BUCKET s3://test-bucket -e testing

./epicenv set LOG_LEVEL info -e agent-testing

# 5. Verify stacking - get secrets from agent-testing (should show stacked values)
./epicenv get DB_HOST -e agent-testing
./epicenv get S3_BUCKET -e agent-testing
./epicenv get LOG_LEVEL -e agent-testing

# 6. Verify local still has its original values
./epicenv get S3_BUCKET -e local
./epicenv get LOG_LEVEL -e local

# 7. Test list-invites on overlay (should warn about root environment)
./epicenv list-invites -e agent-testing

# 8. Test invite on overlay (should warn about adding to root)
# (Skip this if you don't want to add another user)
# ./epicenv invite someuser -e agent-testing

# 9. Test rm on overlay for a key that's in underlay
./epicenv rm DB_HOST -e agent-testing

# 10. Test rm on overlay for a key that exists in this layer
./epicenv rm LOG_LEVEL -e agent-testing
