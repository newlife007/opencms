# Admin Password Reset - Quick Guide

## For System Administrators

### Quick Reset Using Script

The easiest way to reset the admin password:

```bash
cd /home/ec2-user/openwan
./scripts/reset_admin_password.sh [new_password]
```

**Examples:**
```bash
# Reset to default password "admin123"
./scripts/reset_admin_password.sh

# Reset to custom password
./scripts/reset_admin_password.sh myNewPassword123
```

### Manual Reset Using SQL

If you prefer manual control:

1. **Generate password hash:**
```bash
cd /home/ec2-user/openwan
go run gen_admin_password.go
# Output example: Password: admin123
#                 Hash: $2a$10$F/rqoZeP...
```

2. **Update database:**
```sql
USE openwan_db;

-- Update password and enable account
UPDATE ow_users 
SET password = '$2a$10$YOUR_HASH_HERE',
    enabled = 1
WHERE username = 'admin';

-- Verify update
SELECT id, username, enabled, LEFT(password, 30) as pwd_preview 
FROM ow_users 
WHERE username = 'admin';
```

## For End Users

### If You Can't Login After Password Reset

1. **Clear browser data:**
   - Press `Ctrl+Shift+Delete` (Windows/Linux) or `Cmd+Shift+Delete` (Mac)
   - Clear cookies and cached files
   - Or use Incognito/Private browsing mode

2. **Try again:**
   - Go to login page
   - Enter username: `admin`
   - Enter the new password you set
   - Click Login

3. **Still not working?**
   - Make sure you're using the **new** password, not the old one
   - Check if Caps Lock is on
   - Try a different browser
   - Contact your system administrator

## Default Credentials

After running the reset script without parameters:

- **Username:** `admin`
- **Password:** `admin123`
- **Login URL:** http://localhost:3000

## Troubleshooting

### Error: "Invalid username or password"

**Possible causes:**
1. Account is disabled
2. Wrong password
3. Database connection issue

**Solution:**
```bash
# Check account status
mysql -h 127.0.0.1 -u openwan -p'openwan123' openwan_db \
  -e "SELECT id, username, enabled FROM ow_users WHERE username='admin';"

# If enabled = 0, run:
mysql -h 127.0.0.1 -u openwan -p'openwan123' openwan_db \
  -e "UPDATE ow_users SET enabled = 1 WHERE username='admin';"
```

### Error: "Cannot connect to database"

**Solution:**
```bash
# Check if MySQL is running
sudo systemctl status mysql

# If not running:
sudo systemctl start mysql

# Check connection manually
mysql -h 127.0.0.1 -u openwan -p'openwan123' openwan_db -e "SELECT 1;"
```

### Error: "Backend service not responding"

**Solution:**
```bash
# Check if backend is running
curl http://localhost:8080/health

# If not running, start it:
cd /home/ec2-user/openwan
go run cmd/api/main.go
```

## Password Requirements

- **Minimum length:** 6 characters
- **Maximum length:** 32 characters
- **Allowed characters:** Letters, numbers, special characters
- **Case sensitive:** Yes

## Security Recommendations

1. **Change default password immediately** after first login
2. **Use strong passwords** with mix of letters, numbers, and symbols
3. **Don't share passwords** - create separate accounts for each user
4. **Enable two-factor authentication** (future feature)
5. **Regularly update passwords** (every 90 days recommended)

## Common Issues and Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| Login fails after reset | Browser cached old session | Clear cookies and try again |
| "Account disabled" message | `enabled = 0` in database | Run reset script (sets `enabled = 1`) |
| Password not updated | Database connection failed | Check MySQL service status |
| Reset button does nothing in UI | JavaScript error | Check browser console, try different browser |

## Contact Support

If you continue to experience issues:

1. Check `/home/ec2-user/openwan/PASSWORD_RESET_RESOLUTION.md` for detailed analysis
2. Check application logs: `tail -f /var/log/openwan/app.log`
3. Contact system administrator with error details

## Files Reference

- **Reset script:** `/home/ec2-user/openwan/scripts/reset_admin_password.sh`
- **Hash generator:** `/home/ec2-user/openwan/gen_admin_password.go`
- **Detailed docs:** `/home/ec2-user/openwan/PASSWORD_RESET_RESOLUTION.md`
- **Config file:** `/home/ec2-user/openwan/configs/config.yaml`
