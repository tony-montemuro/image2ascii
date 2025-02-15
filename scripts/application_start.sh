#!/bin/bash
LOG_FILE="/var/log/image2ascii/deploy.log"

{
    echo
    echo '=================================================='
    echo
    echo 'Starting application_start.sh: '
    echo 'Starting image2ascii service: '
    systemctl start image2ascii
    echo 'image2ascii service started successfully'
    echo 'application_start.sh completed successfully'
} >> "$LOG_FILE" 2>&1