ROOT_PROJECT	:= gocrack
PKG_AUTHOR		:= Christopher Schmitt <chris.schmitt@mandiant.com>
HASHCAT_VERSION	:= v5.1.0

GOCRACK_PROJ		:= github.com/fireeye/gocrack
GOCRACK_CODE_LOCAL	:= ${GOPATH}/src/$(GOCRACK_PROJ)
EXTERNAL_DIR		:= $(CURDIR)/external
HASHCAT_DIR			:= $(EXTERNAL_DIR)/hashcat
DIST_DIR			:= $(CURDIR)/dist
HASHCAT_DIST		:= $(DIST_DIR)/hashcat
GOCRACK_DIST		:= $(DIST_DIR)/gocrack

DOCKER_BUILD	:= docker build --build-arg USER_ID="$(shell id -u)" \
								--build-arg AUTHOR="$(PKG_AUTHOR)"

DOCKER_RUN		:= docker run --rm --net=host -e USER_ID=$(shell id -u)

hashcat:
		@mkdir -p $(EXTERNAL_DIR)
		@cd $(EXTERNAL_DIR); \
			git clone https://github.com/hashcat/hashcat.git; \
			cd $(HASHCAT_DIR); \
				git submodule update --init; \
				git checkout $(HASHCAT_VERSION);
		$(DOCKER_BUILD) -t $(ROOT_PROJECT)/hashcat_shared -f Dockerfile.hashcat.ubuntu $(CURDIR)
		$(DOCKER_RUN) \
			-v $(HASHCAT_DIST):/out \
			$(ROOT_PROJECT)/hashcat_shared

gocrack:
		@mkdir -p $(GOCRACK_DIST)
		$(DOCKER_BUILD) \
			--build-arg GOCRACK_PROJ="$(GOCRACK_PROJ)" \
			--build-arg HASHCAT_VER="$(HASHCAT_VERSION)" \
			-t $(ROOT_PROJECT)/build \
			-f Dockerfile.build.ubuntu $(CURDIR);
		$(DOCKER_RUN) \
			-v $(GOCRACK_DIST):/out \
			-v $(GOCRACK_CODE_LOCAL):/go/src/$(GOCRACK_PROJ) \
			$(ROOT_PROJECT)/build

build_server_image:
	$(DOCKER_BUILD) -t $(ROOT_PROJECT)/server -f Dockerfile.server.ubuntu $(CURDIR)

build_worker_image:
	$(DOCKER_BUILD) -t $(ROOT_PROJECT)/worker -f Dockerfile.worker.ubuntu $(CURDIR)

build_images: build_server_image build_worker_image

clean:
		@rm -rf $(DIST_DIR)

build: hashcat gocrack build_server_image build_worker_image