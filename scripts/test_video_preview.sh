#!/bin/bash
# Script to test video preview functionality

echo "=== OpenWan Video Preview Testing Script ==="
echo ""

# Check if FFmpeg is installed
echo "1. Checking FFmpeg installation..."
if command -v ffmpeg &> /dev/null; then
    echo "   ✅ FFmpeg found: $(ffmpeg -version | head -1)"
else
    echo "   ❌ FFmpeg not found. Please install FFmpeg."
    exit 1
fi

# Check for video files in database
echo ""
echo "2. Checking for video files in database..."
VIDEO_COUNT=$(sudo docker exec openwan-mysql-1 mysql -uroot -prootpassword openwan_db -se "SELECT COUNT(*) FROM ow_files WHERE type=1;" 2>/dev/null)
echo "   Found $VIDEO_COUNT video files in database"

if [ "$VIDEO_COUNT" -eq 0 ]; then
    echo "   ℹ️  No video files to transcode. Upload a video first."
    exit 0
fi

# Check for existing preview files
echo ""
echo "3. Checking for existing preview files..."
PREVIEW_COUNT=$(find /home/ec2-user/openwan/storage -name "*-preview.flv" 2>/dev/null | wc -l)
echo "   Found $PREVIEW_COUNT preview files"

# Check if RabbitMQ is running
echo ""
echo "4. Checking RabbitMQ status..."
if sudo docker ps | grep -q rabbitmq; then
    echo "   ✅ RabbitMQ container is running"
else
    echo "   ⚠️  RabbitMQ container is not running"
    echo "   Starting RabbitMQ..."
    cd /home/ec2-user/openwan
    sudo docker-compose up -d rabbitmq
    sleep 5
fi

# Check if worker binary exists
echo ""
echo "5. Checking worker application..."
if [ -f "/home/ec2-user/openwan/bin/worker" ]; then
    echo "   ✅ Worker binary exists"
else
    echo "   ⚠️  Worker binary not found. Building..."
    cd /home/ec2-user/openwan
    go build -o bin/worker ./cmd/worker
    if [ $? -eq 0 ]; then
        echo "   ✅ Worker binary built successfully"
    else
        echo "   ❌ Failed to build worker binary"
        exit 1
    fi
fi

# Option to start worker
echo ""
echo "=== Worker Management ==="
echo ""
echo "Options:"
echo "  1) Start worker in background"
echo "  2) Start worker in foreground (with logs)"
echo "  3) Check worker status"
echo "  4) Stop worker"
echo "  5) Manually transcode a file"
echo "  6) Exit"
echo ""
read -p "Select option (1-6): " option

case $option in
    1)
        echo "Starting worker in background..."
        cd /home/ec2-user/openwan
        nohup ./bin/worker > /tmp/worker.log 2>&1 &
        WORKER_PID=$!
        echo $WORKER_PID > /tmp/worker.pid
        echo "   ✅ Worker started with PID: $WORKER_PID"
        echo "   Logs: /tmp/worker.log"
        echo "   To stop: kill $WORKER_PID"
        ;;
    2)
        echo "Starting worker in foreground (Ctrl+C to stop)..."
        cd /home/ec2-user/openwan
        ./bin/worker
        ;;
    3)
        if [ -f "/tmp/worker.pid" ]; then
            WORKER_PID=$(cat /tmp/worker.pid)
            if ps -p $WORKER_PID > /dev/null; then
                echo "   ✅ Worker is running (PID: $WORKER_PID)"
                echo "   Recent logs:"
                tail -20 /tmp/worker.log
            else
                echo "   ❌ Worker is not running (PID file exists but process not found)"
            fi
        else
            echo "   ℹ️  No worker PID file found. Worker may not be running."
        fi
        ;;
    4)
        if [ -f "/tmp/worker.pid" ]; then
            WORKER_PID=$(cat /tmp/worker.pid)
            echo "Stopping worker (PID: $WORKER_PID)..."
            kill $WORKER_PID 2>/dev/null
            rm /tmp/worker.pid
            echo "   ✅ Worker stopped"
        else
            echo "   ℹ️  No worker PID file found"
        fi
        ;;
    5)
        echo ""
        echo "Available video files:"
        sudo docker exec openwan-mysql-1 mysql -uroot -prootpassword openwan_db -e "SELECT id, title, name, ext FROM ow_files WHERE type=1 LIMIT 10;" 2>/dev/null | column -t
        echo ""
        read -p "Enter file ID to transcode: " FILE_ID
        
        # Get file info
        FILE_INFO=$(sudo docker exec openwan-mysql-1 mysql -uroot -prootpassword openwan_db -se "SELECT name, path, ext FROM ow_files WHERE id=$FILE_ID;" 2>/dev/null)
        if [ -z "$FILE_INFO" ]; then
            echo "   ❌ File not found"
            exit 1
        fi
        
        FILE_NAME=$(echo "$FILE_INFO" | cut -f1)
        FILE_PATH=$(echo "$FILE_INFO" | cut -f2)
        FILE_EXT=$(echo "$FILE_INFO" | cut -f3)
        
        # Convert Windows path to Unix path
        FILE_PATH_UNIX=$(echo "$FILE_PATH" | sed 's/\\\\/\//g')
        INPUT_FILE="/home/ec2-user/openwan/storage/${FILE_PATH_UNIX}${FILE_NAME}${FILE_EXT}"
        OUTPUT_FILE="/home/ec2-user/openwan/storage/${FILE_PATH_UNIX}${FILE_NAME}-preview.flv"
        
        echo "   Input:  $INPUT_FILE"
        echo "   Output: $OUTPUT_FILE"
        
        if [ ! -f "$INPUT_FILE" ]; then
            echo "   ❌ Input file not found: $INPUT_FILE"
            exit 1
        fi
        
        echo ""
        echo "Transcoding with FFmpeg..."
        ffmpeg -i "$INPUT_FILE" -y -ab 56 -ar 22050 -r 15 -b:v 500k -s 320x240 "$OUTPUT_FILE" 2>&1 | tail -20
        
        if [ -f "$OUTPUT_FILE" ]; then
            FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
            echo "   ✅ Preview file created: $OUTPUT_FILE ($FILE_SIZE)"
            echo "   You can now view the video at: http://localhost:3000/files/$FILE_ID"
        else
            echo "   ❌ Failed to create preview file"
        fi
        ;;
    6)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

echo ""
echo "=== Testing Complete ==="
echo ""
echo "Next steps:"
echo "  1. Ensure worker is running (option 1 or 2)"
echo "  2. Upload a video file or transcode existing file (option 5)"
echo "  3. Open browser: http://localhost:3000/files/{file_id}"
echo "  4. Check browser console for VideoPlayer logs"
echo ""
