OUTPUT_DIR=${OUTPUT_DIR:-"_output"}

./${OUTPUT_DIR}/conntracker \
 --v=5 \
 --master="http://127.0.0.1:8080" > "/tmp/kube-conntracker.log" 2>&1 &
