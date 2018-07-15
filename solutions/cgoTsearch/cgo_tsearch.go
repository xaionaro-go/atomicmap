//go:generate benchmarkCodeGen

package cgoTsearch

/*
#cgo CFLAGS: -O3
#define _GNU_SOURCE
#include <errno.h>
#include <search.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
struct item {
	unsigned long key;
	void *value;
};
void free_node(void *nodep) {
	free(nodep);
}
void tdestroy_wrapper(void *rootp) { tdestroy(rootp, free_node); }
int key_cmp(const void *itemAV, const void *itemBV) {
	const struct item *itemA = itemAV;
	const struct item *itemB = itemBV;
	return itemA->key - itemB->key;
}
void *tfind_wrapper(unsigned long key, void *const *rootp) {
	struct item item;
	item.key = key;
	struct item **found_item = tfind(&item, rootp, key_cmp);
	if (found_item == NULL) {
		return NULL;
	}
	return &((*found_item)->value);
}
void tsearch_wrapper(unsigned long key, void *value, void **rootp) {
	struct item *item = malloc(sizeof(struct item *));
	item->key = key;
	item->value = value;
	struct item **found_item = tsearch(item, rootp, key_cmp);
	if (item != *found_item) {
		free(item);
	}
}
void tdelete_wrapper(unsigned long key, void **rootp) {
	struct item item;
	item.key = key;
	tdelete(&item, rootp, key_cmp);
}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	"git.dx.center/trafficstars/testJob0/internal/routines"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

type tsearch struct {
	rootp *C.void
}

func freeTsearch(tsearchI interface{}) {
	tsearch := tsearchI.(*tsearch)
	tsearch.destroy()
}

func NewHashMap(blockSize int, fn func(blockSize int, key I.Key) int) I.HashMaper {
	result := &tsearch{}
	runtime.SetFinalizer(result, freeTsearch)
	return result
}

func convertKey(keyI I.Key) C.ulong {
	intKey := routines.HashFunc(1<<31, keyI)
	return C.ulong(intKey)
}

func (m *tsearch) destroy() {
	C.tdestroy_wrapper(unsafe.Pointer(m.rootp))
}
func (m *tsearch) rootpp() *unsafe.Pointer {
	ptr := &m.rootp
	return (*unsafe.Pointer)((unsafe.Pointer)(ptr))
}
func (m *tsearch) Get(key I.Key) (interface{}, error) {
	result := C.tfind_wrapper(convertKey(key), m.rootpp())
	if result == nil {
		return nil, errors.NotFound
	}
	return *((*int)(result)), nil
}
func (m *tsearch) Set(key I.Key, value interface{}) error {
	valueInt := value.(int)
	C.tsearch_wrapper(convertKey(key), unsafe.Pointer(uintptr(valueInt)), m.rootpp())
	return nil
}
func (m *tsearch) Unset(key I.Key) error {
	C.tdelete_wrapper(convertKey(key), m.rootpp())
	return nil
}

func (m tsearch) Count() int {
	return -1
}
func (m tsearch) CheckConsistency() error {
	return nil
}
