BINARY_FOLDER := bin
OS := $(shell uname -s)
FLAGS := ""

all: build-all
build-all: ${OS} ${BINARY_FOLDER}/koru ${BINARY_FOLDER}/korucli ${BINARY_FOLDER}/korued
	@echo Built everything
${BINARY_FOLDER}:
	@echo Create binary folder
	mkdir -p ${BINARY_FOLDER}
${BINARY_FOLDER}/assets: ${BINARY_FOLDER}
	@echo Copying assets
	cp -r assets/ ${BINARY_FOLDER}/
vendor:
	@echo Installing dependencies
	dep ensure
	make fix-vulkan
${BINARY_FOLDER}/koru: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling koru
	cd ./cmd/koru && \
	go build -tags=vulkan -o ../../${BINARY_FOLDER}/koru
${BINARY_FOLDER}/korucli: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling korucli
	cd ./cmd/korucli && \
	go build -o ../../${BINARY_FOLDER}/korucli
${BINARY_FOLDER}/korued: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling korued
	cd ./cmd/korued && \
	packr && \
	go build -o ../../${BINARY_FOLDER}/korued
${BINARY_FOLDER}/shaders: ${BINARY_FOLDER}
	mkdir ${BINARY_FOLDER}/shaders
	./buildShaders.sh ${BINARY_FOLDER}/shaders

Linux:
	@echo Linux specific prepare
Darwin:
	@echo Darwin specific prepare
	
fix-vulkan:
	@echo Regenerate bindings
	cd vendor/github.com/vulkan-go/vulkan && \
	make clean && \
	c-for-go -ccdefs -ccincl -out .. vulkan.yml

clean:
	rm -rf vendor ${BINARY_FOLDER}