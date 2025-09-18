BINARY_NAME=tank
FRONT_DIR=front
BUILD_DIR=$(FRONT_DIR)/build
HTML_DIR=html
MAIN_GO=main.go

all: build

build-front:
	cd $(FRONT_DIR) && yarn build
	rm -rf $(HTML_DIR)
	mkdir $(HTML_DIR)
	cp -R $(BUILD_DIR)/* $(HTML_DIR)/

build-backend:
	@echo "构建后端..."
	go build -o $(BINARY_NAME) $(MAIN_GO)

build: build-front build-backend

run: build
	@echo "运行程序..."
	./$(BINARY_NAME)

clean:
	@echo "清理构建产物..."
	rm -f $(BINARY_NAME)
	rm -rf $(HTML_DIR)/*
	cd $(FRONT_DIR) && yarn clean

dev: build-backend
	@echo "同时启动前端和后端开发服务器..."
	trap 'kill $(jobs -p)' EXIT; \
	(cd $(FRONT_DIR) && yarn start) & \
	(TANK=DEVELOPMENT ./$(BINARY_NAME)) & \
	wait

dev-front:
	cd $(FRONT_DIR) && yarn start

.PHONY: all build-front build-backend build run clean dev-front dev-back dev
    