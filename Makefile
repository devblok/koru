BINARY_FOLDER := "bin"
OS := "$(uname -s)"
FLAGS := ""

all: build-all
build-all: ${OS} ${BINARY_FOLDER} ${BINARY_FOLDER}/koru ${BINARY_FOLDER}/korucli ${BINARY_FOLDER}/korued
	echo Building everything
${BINARY_FOLDER}:
	echo Create binary foler
	mkdir ${BINARY_FOLDER}
vendor:
	echo Installing dependencies
	dep ensure
${BINARY_FOLDER}/koru: vendor
	echo Compiling koru
	pushd ./cmd/koru
	go build -o ../../${BINARY_FOLDER}/koru
	popd
${BINARY_FOLDER}/korucli: vendor
	echo Compiling korucli
	pushd ./cmd/korucli
	go build -o ../../${BINARY_FOLDER}/korucli
	popd
${BINARY_FOLDER}/korued: vendor
	echo Compiling korued
	pushd ./cmd/korued
	go build -o ../../${BINARY_FOLDER}/korued
	popd

Linux:
Darwin:
	FLAGS := ""

clean:
	rm -rf vendor ${BINARY_FOLDER}