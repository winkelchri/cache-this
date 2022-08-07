ifeq ($(OS),Windows_NT)
	include Makefile_windows
else
	include Makefile_linux
endif
