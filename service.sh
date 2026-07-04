#!/system/bin/sh
   MODDIR=${0%/*}
   chmod -R 755 "$MODDIR"
   find "$MODDIR" -type f \( -name '*.prop' -o -name '*.json' -o -name '*.log' -o -name '*.md' -o -name '*.txt' \) -exec chmod 644 {} + 2>/dev/null
   "$MODDIR/service" > /dev/null 2>&1 &
   pkill -f timer_daemon
   sleep 3
   setsid "$MODDIR/cron/timer_daemon" > /dev/null 2>&1 &