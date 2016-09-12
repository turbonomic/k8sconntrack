CONN_PID=`/bin/ps -fu $USER| grep "conntracker " | grep -v "grep" | awk '{print $2}'`
echo $CONN_PID
kill $CONN_PID
