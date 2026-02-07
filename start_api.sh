#!/bin/bash
cd /home/ec2-user/openwan
./bin/openwan-api --config configs/config.yaml >> logs/api.log 2>&1 &
echo $! > /tmp/openwan-api.pid
echo "OpenWan API started with PID $(cat /tmp/openwan-api.pid)"
