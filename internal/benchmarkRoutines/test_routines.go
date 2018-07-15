package benchmarkRoutines

import (
	"testing"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	"git.dx.center/trafficstars/testJob0/internal/routines"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
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

func DoTest(t *testing.T, factoryFunc mapFactoryFunc) {
	m := factoryFunc(1024, routines.HashFunc)

	if m.Count() != 0 {
		t.Errorf("m.Count() is not 0: %v", m.Count())
	}

	m.Set(1024*1024, 1)
	m.Set("a string", 2)
	m.Set([]byte{1, 2, 3}, 3)
	m.Set(map[string]string{"hello": "world"}, 4)

	expect(t, m, 1024*1024, 1)
	expect(t, m, "a string", 2)
	expect(t, m, []byte{1, 2, 3}, 3)
	expect(t, m, map[string]string{"hello": "world"}, 4)

	_, err := m.Get(3)
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	_, err = m.Get([]byte{1, 2, 3, 0})
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	_, err = m.Get([]byte{0, 1, 2, 3})
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	_, err = m.Get([]byte("a string"))
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	if m.Count() != 4 {
		t.Errorf("m.Count() is not 4: %v", m.Count())
	}

	err = m.Unset(1024*1024)
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}

	_, err = m.Get(1024*1024)
	if err != errors.NotFound {
		t.Errorf(`An expected "NotFound" error, but got: %v`, err)
	}

	if m.Count() != 3 {
		t.Errorf("m.Count() is not 3: %v", m.Count())
	}

	for i:=10; i<1024*128; i++ {
		m.Set(i, i)
	}

	err = m.(checkConsistencier).CheckConsistency()
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
		return
	}

	for i:=0; i<10; i++ {
		m.Set(i, i)
	}

	err = m.(checkConsistencier).CheckConsistency()
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}
}
