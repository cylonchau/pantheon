.PHONY: all build run gotool clean help
.DEFAULT_GOAL := help

GOCMD=go
GOBUILD=$(GOCMD) build
GOBUILD_DIR=cmd
OUT_DIR ?= _output
BIN_DIR := $(OUT_DIR)/bin

# 定义目标
modules := $(wildcard $(GOBUILD_DIR)/*)

# 从子目录中提取目标名称
SUBDIRS := $(notdir $(modules))

build:
	@if [ -z "$(module)" ]; then \
		echo "No module specified."; \
		echo "Usage: make build module=<subdir>"; \
		echo "Available modules are: $(SUBDIRS)"; \
		exit 0; \
	fi
	@if echo "$(SUBDIRS)" | grep -qw "$(module)"; then \
    	scripts/build.sh $(module); \
	else \
		echo "Error: Invalid module '$(module)'. Available modules are: $(SUBDIRS)"; \
		exit 1; \
	fi

clean:
	@if [ -d "$(OUT_DIR)" ]; then \
		rm -rf $(OUT_DIR); \
	fi

help:
	@echo "Available commands:"
	@echo "  make build module=<subdir>  # Build the specified module"
	@echo "  make clean                   # Clean output directory"
	@echo "  make help                    # Show this help message"
	@echo ""
	@echo "Example:"
	@for dir in $(SUBDIRS); do \
		echo "  make build module=$$dir"; \
	done
