#!/bin/bash
LOG_FILE="/var/log/image2ascii/deploy.log"

{
    # Begin script
    echo 'Starting after_install.sh: '
    cd /home/ubuntu/image2ascii
    echo 'Successfully changed directory to /home/ubuntu/image2ascii'
    
    # Update dependencies
    echo 'Updating go dependencies...'
    go mod tidy
    echo 'Updating js dependencies...'
    npm install

    # Building + compiling
    echo 'Rebuilding output.css file...'
    npx tailwindcss -i ./static/input.css -o ./static/styles.css
    echo 'Recompiling go binary...'
    go build

    echo 'after_install.sh completed successfully'

} >> "$LOG_FILE" 2>&1