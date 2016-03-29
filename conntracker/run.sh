./conntracker \
 --v=3 \
 --node-addr="127.0.0.1" \
 --master="http://127.0.0.1:8080" > "/tmp/kube-conntracker.log" 2>&1 &
