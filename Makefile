BINARY_FOLDER := bin
OS := $(shell uname -s)
#FLAGS := -compiler gccgo -gccgoflags "-Os -O2"
FLAGS := ""

COL=\033[1;33m # Stage color - yellow
SCOL=\033[0;32m # Success color - green
NC=\033[0m # No Color

all: build-all
build-all: ${OS} koru korucli korued test-all
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
vendor:
	@printf "${COL}Installing dependencies${NC}\n"
	dep ensure
	make fix-vulkan
koru: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling koru${NC}\n"
	cd ./cmd/koru && \
	go build -tags=vulkan -o ../../${BINARY_FOLDER}/koru ${FLAGS}
korucli: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling korucli${NC}\n"
	cd ./cmd/korucli && \
	go build -o ../../${BINARY_FOLDER}/korucli ${FLAGS}
korued: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@printf "${COL}Compiling korued${NC}\n"
	cd ./cmd/korued && \
	packr && \
	go build -o ../../${BINARY_FOLDER}/korued ${FLAGS}
${BINARY_FOLDER}/shaders: ${BINARY_FOLDER}
	mkdir -p ${BINARY_FOLDER}/shaders
	./buildShaders.sh ${BINARY_FOLDER}/shaders

Linux:
	@printf "${COL}Linux specific prepare${NC}\n"
Darwin:
	@printf "${COL}Darwin specific prepare${NC}\n"
	
fix-vulkan:
	@printf "${COL}Regenerate bindings${NC}\n"
	cd vendor/github.com/vulkan-go/vulkan && \
	make clean && \
	c-for-go -ccdefs -ccincl -out .. vulkan.yml

test-all: test-unit
test-unit:
	@printf "${COL}Running unit tests${NC}\n"
	go test ./...

clean:
	rm -rf vendor ${BINARY_FOLDER}