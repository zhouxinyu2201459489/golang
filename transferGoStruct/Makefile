# 设置编译器路径和编译参数
GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=transferGoStruct

# 定义默认构建和安装任务
all: build install

# 定义构建任务
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# 定义安装任务
install:
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# 定义清理任务
clean:
	rm -f $(BINARY_NAME)
