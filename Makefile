OUTPUT_DIR=./_output
BINARY=${OUTPUT_DIR}/conntracker

build: clean
	go build -o ${BINARY} ./cmd

.PHONY: clean
clean:
	@: if [ -f ${OUTPUT_DIR} ]; then rm -rf ${OUTPUT_DIR};fi
