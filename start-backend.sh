#!/bin/bash

# OpenWan Backend启动脚本
# 设置存储路径环境变量

cd /home/ec2-user/openwan

# 设置环境变量
export LOCAL_STORAGE_PATH=/home/ec2-user/openwan/data

# 启动backend
./bin/openwan
