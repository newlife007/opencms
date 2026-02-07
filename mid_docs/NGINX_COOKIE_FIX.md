# Nginx Cookie Forwarding Fix

## Problem
File download failed with HTTP 500 error when accessed through nginx proxy:
- URL: `http://13.217.210.142/api/v1/files/72/download`
- Error: 500 (Internal Server Error)
- Root Cause: Nginx was not forwarding cookies to the backend API

## Root Causes Identified

### 1. Missing Cookie Forwarding in Nginx
The nginx configuration was not passing cookies from the client to the backend Go application, causing session authentication to fail.

### 2. Empty Nginx Configuration File
During troubleshooting, the `openwan.conf` file was accidentally created as empty (0 bytes), causing nginx to not listen on port 80.

## Solutions Applied

### Fix 1: Added Cookie Forwarding Headers

Updated `/etc/nginx/conf.d/openwan.conf` in the `/api/` location block:

```nginx
location /api/ {
    proxy_pass http://localhost:8080/api/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # Cookie forwarding for session authentication
    proxy_set_header Cookie $http_cookie;
    proxy_pass_header Set-Cookie;
    
    # Timeouts for large file downloads/uploads
    proxy_connect_timeout 300s;
    proxy_send_timeout 600s;
    proxy_read_timeout 600s;
    proxy_request_buffering off;
}
```

### Fix 2: Restored Nginx Configuration

Restored the working configuration from backup:
```bash
sudo cp /etc/nginx/conf.d/openwan.conf.bak /etc/nginx/conf.d/openwan.conf
```

### Fix 3: Restarted Nginx

```bash
sudo nginx -t
sudo systemctl restart nginx
```

## Verification

### Test 1: Login Through Nginx
```bash
curl -c /tmp/cookies.txt -X POST http://13.217.210.142/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test123"}'
```
**Result**: ✅ Success - `{"success":true}`

### Test 2: File Download Through Nginx
```bash
curl -b /tmp/cookies.txt \
  http://13.217.210.142/api/v1/files/72/download \
  -o /tmp/file72.png
```
**Result**: ✅ Success - HTTP 200, 69 bytes PNG image downloaded

### Test 3: Verify File Integrity
```bash
file /tmp/file72.png
# Output: PNG image data, 1 x 1, 8-bit/color RGB, non-interlaced
```
**Result**: ✅ File integrity confirmed

## Key Configuration Changes

| Component | Before | After |
|-----------|--------|-------|
| Cookie Forwarding | ❌ Not configured | ✅ `proxy_set_header Cookie $http_cookie` |
| Set-Cookie Passthrough | ❌ Not configured | ✅ `proxy_pass_header Set-Cookie` |
| Port 80 Listening | ❌ Not listening | ✅ Listening on 0.0.0.0:80 |

## Impact

- **File Downloads**: ✅ Now working through nginx proxy with authentication
- **Session Authentication**: ✅ Cookies properly forwarded to backend
- **Frontend Access**: ✅ All API calls through http://13.217.210.142/api/* now work correctly

## Files Modified

1. `/etc/nginx/conf.d/openwan.conf` - Added cookie forwarding headers
2. Restored from backup: `/etc/nginx/conf.d/openwan.conf.bak`

## Testing Checklist

- [x] Login through nginx (http://13.217.210.142/api/v1/auth/login)
- [x] Check authenticated endpoint (http://13.217.210.142/api/v1/auth/me)
- [x] File download with authentication (http://13.217.210.142/api/v1/files/72/download)
- [x] Verify file integrity after download
- [x] Nginx listening on port 80
- [x] Backend receiving cookies

## Status

✅ **RESOLVED** - File download through nginx proxy now works correctly with session authentication.

## Date
2026-02-03 03:22 UTC
