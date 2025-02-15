#!/bin/bash
LOG_FILE="/var/log/image2ascii/deploy.log"

{
    echo
    echo '=================================================='
    echo
    echo 'Starting before_install.sh: '
    echo 'Stopping image2ascii service: '
    systemctl stop image2ascii
    echo 'image2ascii service stopped successfully'
    echo 'before_install.sh completed successfully'
} >> "$LOG_FILE" 2>&1