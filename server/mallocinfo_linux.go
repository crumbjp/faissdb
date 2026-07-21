//go:build linux

package main

/*
#include <malloc.h>
#include <stdio.h>
#include <stdlib.h>

static char* go_faissdb_malloc_info() {
	char* buffer = NULL;
	size_t size = 0;
	FILE* stream = open_memstream(&buffer, &size);
	if (stream == NULL) {
		return NULL;
	}
	if (malloc_info(0, stream) != 0) {
		fclose(stream);
		free(buffer);
		return NULL;
	}
	fclose(stream);
	return buffer;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

func MallocInfo() (string, error) {
	buffer := C.go_faissdb_malloc_info()
	if buffer == nil {
		return "", errors.New("MallocInfo() malloc_info failed")
	}
	defer C.free(unsafe.Pointer(buffer))
	return C.GoString(buffer), nil
}
