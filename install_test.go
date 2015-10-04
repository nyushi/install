package install

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
)

var (
	dir string
)

func TestInstallFile(t *testing.T) {
	src := path.Join(dir, "1")
	originalData := []byte(`this is test file`)
	ioutil.WriteFile(src, originalData, 0644)

	dst := path.Join(dir, "2")

	if err := InstallFile(src, dst, nil); err != nil {
		t.Fatalf("install error: %s", err)
	}

	data, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst file")
	}
	if !bytes.Equal(data, originalData) {
		t.Fatalf("mismatch: got: %s, expected: %s", data, originalData)
	}
}

func TestInstallFileToDir(t *testing.T) {
	srcName := "3"
	src := path.Join(dir, srcName)
	originalData := []byte(`this is test file`)
	ioutil.WriteFile(src, originalData, 0600)

	dstDir := path.Join(dir, "4")
	os.Mkdir(dstDir, 0755)

	if err := InstallFile(src, dstDir, nil); err != nil {
		t.Fatalf("install error: %s", err)
	}

	data, err := ioutil.ReadFile(path.Join(dstDir, srcName))
	if err != nil {
		t.Fatalf("failed to read dst file")
	}
	if !bytes.Equal(data, originalData) {
		t.Fatalf("mismatch: got: %s, expected: %s", data, originalData)
	}
}

func TestInstallDir(t *testing.T) {
	dir := path.Join(dir, "5")
	if err := InstallDir(dir, nil); err != nil {
		t.Fatalf("install error: %s", err)
	}

	exists, isDir, err := checkPath(dir)
	if err != nil {
		t.Fatalf("check error: %s", err)
	}
	if !(exists && isDir) {
		t.Fatalf("install failed")
	}
}

func TestMain(m *testing.M) {
	dir = path.Join(os.TempDir(), fmt.Sprintf("install_test_%d", os.Getpid()))
	if err := os.Mkdir(dir, 0777); err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	if err := os.RemoveAll(dir); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
