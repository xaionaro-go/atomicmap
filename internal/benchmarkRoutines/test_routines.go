package benchmarkRoutines

import (
	"fmt"
	"testing"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

const (
	collisionCheckIterations = 1 << 20
)

type checkConsistencier interface {
	CheckConsistency() error
}

func expect(t *testing.T, m I.HashMaper, key I.Key, expectedValue int) {
	value, err := m.Get(key)
	if err != nil {
		t.Errorf("Got an unexpected error: %v. key == %v; expectedValue == %v", err, key, expectedValue)
		return
	}
	if value != expectedValue {
		t.Errorf(`A wrong value "%v" (instead of %v)`, value, expectedValue)
	}
}

func DoTest(t *testing.T, factoryFunc mapFactoryFunc, hashFunc hashFunc) {
	m := factoryFunc(1024, hashFunc)

	if m.Count() != 0 && m.Count() != -1 { // "-1" means "unsupported"
		t.Errorf("m.Count() is not 0: %v", m.Count())
	}

	m.Set(1024*1024, 1)
	m.Set("a string", 2)

	expect(t, m, 1024*1024, 1)
	expect(t, m, "a string", 2)

	_, err := m.Get(3)
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	if m.Count() != 2 && m.Count() != -1 { // "-1" means "unsupported"
		t.Errorf("m.Count() is not 2: %v", m.Count())
	}

	err = m.Unset(1024 * 1024)
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}

	_, err = m.Get(1024 * 1024)
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	if m.Count() != 1 && m.Count() != -1 { // "-1" means "unsupported"
		t.Errorf("m.Count() is not 1: %v", m.Count())
	}

	for i := 10; i < 1024*128; i++ {
		m.Set(i, i)
	}

	err = m.(checkConsistencier).CheckConsistency()
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
		return
	}

	for i := 0; i < 10; i++ {
		m.Set(i, i)
	}

	err = m.(checkConsistencier).CheckConsistency()
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}
}

func DoTestCollisions(t *testing.T, factoryFunc mapFactoryFunc, hashFunc hashFunc) {
	m := factoryFunc(16*collisionCheckIterations, hashFunc)
	keys := generateKeys(collisionCheckIterations/2, "int")
	keys = append(keys, generateKeys(collisionCheckIterations/2, "string")...)

	collisions := 0
	for _, key := range keys {
		if m.(interface{ HasCollisionWithKey(I.Key) bool }).HasCollisionWithKey(key) {
			collisions++
		}
		m.Set(key, true)
	}

	fmt.Printf("Total collisions: %v/%v (%.1f%%)\n", collisions, collisionCheckIterations, float32(collisions)*100/float32(collisionCheckIterations))
}

func DoTestHashCollisions(t *testing.T, hashFunc hashFunc, blockSize uint32, keyAmount uint64) {
	keys := generateKeys(keyAmount, "int")
	keys = append(keys, generateKeys(keyAmount/2, "string")...)

	alreadyIsSet := map[int]bool{}

	collisions := 0
	for _, key := range keys {
		newHash := hashFunc(int(blockSize), key)
		if alreadyIsSet[newHash] {
			collisions++
		}
		alreadyIsSet[newHash] = true
	}

	fmt.Printf("Total collisions on keyAmount %v and blockSize %v: %v/%v (%.1f%%)\n", keyAmount, blockSize, collisions, keyAmount, float32(collisions)*100/float32(keyAmount))
}
