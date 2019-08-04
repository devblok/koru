BINARY_FOLDER := bin
OS := $(shell uname -s)
#FLAGS := -compiler gccgo -gccgoflags "-Os -O2"
FLAGS := ""

all: build-all
build-all: ${OS} koru korucli korued test-all
	@echo Built everything
${BINARY_FOLDER}/koru: koru
${BINARY_FOLDER}/korucli: korucli
${BINARY_FOLDER}/korued: korued
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
koru: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling koru
	cd ./cmd/koru && \
	go build -tags=vulkan -o ../../${BINARY_FOLDER}/koru ${FLAGS}
korucli: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling korucli
	cd ./cmd/korucli && \
	go build -o ../../${BINARY_FOLDER}/korucli ${FLAGS}
korued: vendor ${BINARY_FOLDER} ${BINARY_FOLDER}/assets ${BINARY_FOLDER}/shaders
	@echo Compiling korued
	cd ./cmd/korued && \
	packr && \
	go build -o ../../${BINARY_FOLDER}/korued ${FLAGS}
${BINARY_FOLDER}/shaders: ${BINARY_FOLDER}
	mkdir -p ${BINARY_FOLDER}/shaders
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

test-all: test-unit
test-unit:
	go test ./...

clean:
	rm -rf vendor ${BINARY_FOLDER}