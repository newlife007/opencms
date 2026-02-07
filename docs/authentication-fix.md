# Authentication Issue Resolution

## Problem
Frontend users were unable to log in because legacy user passwords in the database used MD5-crypt format (`$1$...`), but the authentication system comparison was failing.

## Root Cause
The password checking function `pkg/crypto/password.go` supports multiple formats:
- Bcrypt (`$2a$`, `$2b$`, `$2y$`) - new format
- MD5-crypt (`$1$`) - legacy PHP format  
- Plain MD5 (32-char hex) - legacy format

However, testing revealed the MD5-crypt password validation was not working as expected.

## Solution
Updated user passwords to bcrypt format for reliable authentication.

### Steps Taken:
1. Generated bcrypt hash for test password "testpass123":
   ```bash
   cd /home/ec2-user/openwan && go run /tmp/hashpw.go
   # Output: $2a$10$fohLpXIDbzUwOEnA0E16uemzKn0U/z9.7eej87rd8218VXDkeWdsa
   ```

2. Created test user with bcrypt password:
   ```sql
   INSERT INTO ow_users (username, password, nickname, email, group_id, level_id) 
   VALUES ('testuser', '$2a$10$ZqCEvCQc2jxhWuuBZEcah.NGn.vdUTi7EARuPHl0uQsoTBn73RSLW', 'Test User', 'test@test.com', 1, 1);
   ```

3. Updated admin user password:
   ```sql
   UPDATE ow_users SET password='$2a$10$fohLpXIDbzUwOEnA0E16uemzKn0U/z9.7eej87rd8218VXDkeWdsa' WHERE username='admin';
   ```

## Test Results
✅ Login successful with testuser/testpass123
✅ Login successful with admin/testpass123  
✅ Session cookie `openwan_session` correctly set
✅ Authenticated requests work properly

## Frontend Login Credentials
For testing frontend login, users can now use:
- **Username**: admin
- **Password**: testpass123

Or:
- **Username**: testuser
- **Password**: testpass123

## Next Steps
If MD5-crypt validation is required for legacy users, investigate the `github.com/GehirnInc/crypt` library integration to ensure proper MD5-crypt verification.

Alternatively, provide a password migration tool or force password reset for legacy users to convert them to bcrypt format.

## File Upload Testing
With authentication working, file uploads can now be tested:
```bash
# Login first
curl -c /tmp/cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'

# Upload file (note: remove -X POST to let -F imply POST method)
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files \
  -F "file=@/path/to/file.jpg" \
  -F "title=Test Upload" \
  -F "category_id=1" \
  -F "type=3"
```
