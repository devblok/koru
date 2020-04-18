BINARY_FOLDER := bin
OS := $(shell uname -s)
#FLAGS := -compiler gccgo -gccgoflags "-Os -O2"
FLAGS := ""

COL=\033[1;33m # Stage color - yellow
SCOL=\033[0;32m # Success color - green
NC=\033[0m # No Color

all: build-all
build-all: ${OS} kar koru korucli korued test-all
	@printf "${SCOL}Built everything${NC}\n"
${BINARY_FOLDER}/koru: koru
${BINARY_FOLDER}/korucli: korucli
${BINARY_FOLDER}/korued: korued
${BINARY_FOLDER}:
	@printf "${COL}Create binary folder${NC}\n"
	mkdir -p ${BINARY_FOLDER}
${BINARY_FOLDER}/assets: ${BINARY_FOLDER}
	@printf "${COL}Copying assets${NC}\n"
	cp -r assets/ ${BINARY_FOLDER}/
dependencies:
	@printf "${COL}Installing dependencies${NC}\n"
	go mod download
kar: dependencies ${BINARY_FOLDER}
	@printf "${COL}Compiling kar${NC}\n"
	cd ./src/cmd/kar && \
	go build -o ../../../${BINARY_FOLDER}/kar ${FLAGS}
koru: dependencies ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling koru${NC}\n"
	cd ./src/cmd/koru && \
	go build -tags=vulkan -o ../../../${BINARY_FOLDER}/koru ${FLAGS}
korucli: dependencies ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling korucli${NC}\n"
	cd ./src/cmd/korucli && \
	go build -o ../../../${BINARY_FOLDER}/korucli ${FLAGS}
korued: dependencies ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling korued${NC}\n"
	cd ./src/cmd/korued && \
	packr && \
	go build -o ../../../${BINARY_FOLDER}/korued ${FLAGS}
${BINARY_FOLDER}/shaders: ${BINARY_FOLDER}
	mkdir -p ${BINARY_FOLDER}/shaders
	./buildShaders.sh ${BINARY_FOLDER}/shaders

Linux:
	@printf "${COL}Linux specific prepare${NC}\n"
Darwin: dependencies
	@printf "${COL}Darwin specific prepare${NC}\n"

test-all: test-unit benchmark
test-unit:
	@printf "${COL}Running unit tests${NC}\n"
	go test ./...

benchmark: benchmark-core benchmark-model benchmark-kar
benchmark-core:
	@printf "${COL}Benchmarking core package${NC}\n"
	cd ./src/core && go test -bench .
benchmark-model:
	@printf "${COL}Benchmarking model package${NC}\n"
	cd ./src/model && go test -bench .
benchmark-kar:
	@printf "${COL}Benchmarking kar package${NC}\n"
	cd ./src/utility/kar && go test -bench .
clean:
	rm -rf ${BINARY_FOLDER}