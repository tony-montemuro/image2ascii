#!/bin/bash
LOG_FILE="/var/log/image2ascii/deploy.log"

{
    echo 'Starting application_stop.sh: '
    echo 'Stopping image2ascii service: '
    systemctl stop image2ascii
    echo 'image2ascii service stopped successfully'
    echo 'application_stop.sh completed successfully'
} >> "$LOG_FILE" 2>&1