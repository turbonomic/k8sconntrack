OUTPUT_DIR=${OUTPUT_DIR:-"_output"}

./${OUTPUT_DIR}/conntracker \
 --v=3 \
 --master="http://172.17.0.1:8080" \
 --conntrack-port="2223" > "/tmp/kube-conntracker.log" 2>&1 &
