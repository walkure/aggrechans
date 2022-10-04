#!/bin/sh

if [ "${SLACK_BOT_TOKEN}"x == "x" ]; then
    echo "SLACK_BOT_TOKEN required.exit"
    exit 2
fi

if [ "${SLACK_APP_TOKEN}"x != "x" ]; then
    echo "SLACK_APP_TOKEN found. start socket mode"
    exec /app/socket
else
    echo "SLACK_APP_TOKEN not found. start webhook mode"
    exec /app/webhook
fi
