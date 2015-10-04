package install

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"

	"github.com/naegelejd/go-acl/os/group"
)

type InstallOption struct {
	Owner string
	Group string
	UID   string
	GID   string
	Mode  *os.FileMode
}

func (opt *InstallOption) UIDInt() int {
	i, _ := strconv.Atoi(opt.UID)
	return i
}

func (opt *InstallOption) GIDInt() int {
	i, _ := strconv.Atoi(opt.GID)
	return i
}

func (opt *InstallOption) Prepare() error {
	if opt.Owner != "" {
		if opt.UID != "" {
			return errors.New("conflict: Owner and UID")
		}
		u, err := user.Lookup(opt.Owner)
		if err != nil {
			return fmt.Errorf("failed to lookup user `%s`: %s", opt.Owner, err)
		}
		opt.UID = u.Gid
	}
	if opt.Group != "" {
		if opt.GID != "" {
			return errors.New("conflict: Group and GID")
		}
		g, err := group.Lookup(opt.Group)
		if err != nil {
			return fmt.Errorf("failed to lookup group `%s`: %s", opt.Group, err)
		}
		opt.GID = g.Gid
	}

	if opt.UID != "" && opt.GID != "" {
		return nil
	}

	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %s", err)
	}

	if opt.UID == "" {
		opt.UID = u.Uid
	}
	if opt.GID == "" {
		opt.GID = u.Gid
	}
	return nil
}

func tmpPath(s string) string {
	return fmt.Sprintf("%s.install.%d", s, os.Getpid())
}

func InstallFile(src, d string, opt *InstallOption) error {
	if opt == nil {
		opt = &InstallOption{}
	}
	if err := opt.Prepare(); err != nil {
		return fmt.Errorf("invalid option:%s", err)
	}
	dst := d

	dstInfo, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat %s: %s", dst, err)
		}
	}

	if dstInfo != nil {
		// dst exists
		if dstInfo.IsDir() {
			dst = path.Join(dst, filepath.Base(src))
		}
	}

	tmpDst := tmpPath(dst)

	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to read src file(%s): %s", src, err)
	}
	defer r.Close()

	w, err := os.Create(tmpDst)
	if err != nil {
		return fmt.Errorf("failed to create temporary file(%s): %s", tmpDst, err)
	}
	defer w.Close()
	defer os.RemoveAll(tmpDst)

	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)
	if _, err := io.Copy(bw, br); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %s", src, tmpDst, err)
	}
	bw.Flush()

	uid := opt.UIDInt()
	gid := opt.GIDInt()
	if err := w.Chown(uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s %d:%d: %s", tmpDst, uid, gid, err)
	}

	if opt.Mode != nil {
		if err := w.Chmod(*opt.Mode); err != nil {
			return fmt.Errorf("failed to chmod %s %o: %s", tmpDst, *opt.Mode, err)
		}
	}

	if err := os.Rename(tmpDst, dst); err != nil {
		return fmt.Errorf("failed to move %s to %s: %s", tmpDst, dst, err)
	}
	return nil
}

func checkPath(name string) (exists, isDir bool, err error) {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}
	exists = true
	if fi.IsDir() {
		isDir = true
	}
	return
}

func mkdirRec(name string, createMode os.FileMode, createUID, createGID int) error {
	exists, isDir, err := checkPath(name)
	if err != nil {
		return err
	}

	if exists {
		if !isDir {
			return fmt.Errorf("%s exists and not a directory", name)
		}
		return nil
	}

	parent := filepath.Dir(name)
	if err := mkdirRec(parent, createMode, createUID, createGID); err != nil {
		return err
	}

	if err := os.Mkdir(name, createMode); err != nil {
		return fmt.Errorf("error at mkdir for %s: %s", name, err)
	}
	if err := os.Chown(name, createUID, createGID); err != nil {
		return fmt.Errorf("error at chown for:%s: %s", name, err)
	}
	return nil
}

var defaultModeDir = os.FileMode(0755)

func InstallDir(dst string, opt *InstallOption) error {
	if opt == nil {
		opt = &InstallOption{}
	}
	if err := opt.Prepare(); err != nil {
		return fmt.Errorf("invalid option:%s", err)
	}

	if opt.Mode == nil {
		opt.Mode = &defaultModeDir
	}

	if err := mkdirRec(filepath.Dir(dst), *opt.Mode, opt.UIDInt(), opt.GIDInt()); err != nil {
		return fmt.Errorf("failed to create directory: %s", err)
	}

	exists, isDir, err := checkPath(dst)
	if err != nil {
		return fmt.Errorf("failed to check %s: %s", dst, err)
	}
	if exists && !isDir {
		return fmt.Errorf("%s exists and not a directory", dst)
	}

	if !exists {
		if err := os.Mkdir(dst, *opt.Mode); err != nil {
			return fmt.Errorf("error at mkdir for %s: %s", dst, err)
		}
	}

	uid := opt.UIDInt()
	gid := opt.GIDInt()
	if err := os.Chown(dst, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s %d:%d: %s", dst, uid, gid, err)
	}

	if opt.Mode != nil {
		if err := os.Chmod(dst, *opt.Mode); err != nil {
			return fmt.Errorf("failed to chmod %s %o: %s", dst, *opt.Mode, err)
		}
	}
	return nil
}
