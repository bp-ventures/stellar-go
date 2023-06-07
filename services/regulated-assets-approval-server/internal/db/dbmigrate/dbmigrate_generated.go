// Code generated by go-bindata. DO NOT EDIT.
// sources:
// migrations/2021-05-05.0.initial.sql (162B)
// migrations/2021-05-18.0.accounts-kyc-status.sql (414B)
// migrations/2021-06-08.0.pending-kyc-status.sql (193B)
// migrations/2023-06-06.0.fx-rates.sql (335B)

package dbmigrate

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _migrations202105050InitialSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\xcc\xd1\x0d\xc2\x30\x0c\x04\xd0\xff\x4c\x71\xff\x28\x4c\xc1\x08\x30\x80\x01\xa7\xb5\xd4\xda\x91\x6d\xa8\xb2\x3d\x8a\xf8\x40\x7c\xde\xdd\xd3\xd5\x8a\xeb\x2a\x81\x5d\x16\xa7\x14\x53\x34\xd9\x18\x12\x10\x4d\xd6\xd9\xd0\xb6\x0d\xf0\xde\x73\x80\xf4\x39\x27\x42\x13\x8f\x44\x24\x79\x8a\x2e\xe8\x26\x9a\x68\xe6\xa5\x56\xd8\xcb\x7f\x77\x81\x3b\x37\x73\xc6\xc1\x18\x9c\x58\xe9\xcd\x20\xc4\x63\xe5\x9d\xce\x65\xfa\xd3\x17\x33\x6e\xfd\x3f\x5f\xec\xd0\x52\x3e\x01\x00\x00\xff\xff\xd3\x79\x21\xda\xa2\x00\x00\x00")

func migrations202105050InitialSqlBytes() ([]byte, error) {
	return bindataRead(
		_migrations202105050InitialSql,
		"migrations/2021-05-05.0.initial.sql",
	)
}

func migrations202105050InitialSql() (*asset, error) {
	bytes, err := migrations202105050InitialSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/2021-05-05.0.initial.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xd1, 0xd1, 0x21, 0xe9, 0x6d, 0xe0, 0xfe, 0xb4, 0x8b, 0x78, 0x2, 0xae, 0x5c, 0xd5, 0x8b, 0x41, 0xb8, 0x4b, 0xaa, 0x3a, 0xea, 0x69, 0xf, 0xf3, 0x2f, 0x6c, 0xae, 0x38, 0x46, 0xb, 0x2, 0xfc}}
	return a, nil
}

var _migrations202105180AccountsKycStatusSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x90\xc1\x4e\x83\x40\x10\x86\xef\xfb\x14\xff\xb1\x8d\xd6\x17\xe8\x09\x05\x13\x23\x42\x43\x20\xa6\x27\x32\x2c\x13\x5d\xbb\x0b\x9b\xdd\xc1\xaa\x4f\x6f\x02\x26\xda\x13\x1e\x27\xf3\xfd\xdf\x4c\xfe\xdd\x0e\x57\xce\xbc\x04\x12\x46\xe3\x95\xba\xab\xb2\xa4\xce\x50\x27\xb7\x79\x06\x3f\x75\xd6\xe8\x1b\xd2\x7a\x9c\x06\x89\xed\xe9\x53\xb7\x51\x48\xa6\x88\x8d\x02\x80\x28\x6c\x2d\x85\x96\xfa\x3e\x70\x8c\x10\xfe\x10\x14\x65\x8d\xa2\xc9\x73\x1c\xaa\x87\xa7\xa4\x3a\xe2\x31\x3b\x5e\xcf\xb8\x26\x6b\x3b\xd2\xa7\xd6\xf4\x97\xe8\xb2\x66\x47\xc6\x5e\xb8\x7e\x62\x81\x49\xb8\x6f\x49\x20\xc6\x71\x14\x72\x1e\x67\x23\xaf\xf3\x88\xaf\x71\xe0\xdf\xa3\x69\x76\x9f\x34\x79\x8d\xa2\x7c\xde\x6c\x97\xfc\xfc\xf6\xd4\x39\x23\x2b\x96\x05\x27\xef\xc3\xf8\xfe\x1f\x32\xf0\x1b\xeb\x15\xa7\xda\xee\x95\xfa\xdb\x72\x3a\x9e\x07\xa5\xd2\xaa\x3c\xac\xb6\xbc\xff\x0e\x00\x00\xff\xff\x68\xde\x80\x57\x9e\x01\x00\x00")

func migrations202105180AccountsKycStatusSqlBytes() ([]byte, error) {
	return bindataRead(
		_migrations202105180AccountsKycStatusSql,
		"migrations/2021-05-18.0.accounts-kyc-status.sql",
	)
}

func migrations202105180AccountsKycStatusSql() (*asset, error) {
	bytes, err := migrations202105180AccountsKycStatusSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/2021-05-18.0.accounts-kyc-status.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xb0, 0x7b, 0x8c, 0x97, 0xe7, 0x6, 0x27, 0x5f, 0x19, 0xe2, 0xbb, 0x98, 0x73, 0x1e, 0x37, 0x74, 0xf0, 0x4a, 0x7, 0xe7, 0x15, 0x66, 0x90, 0x3c, 0x2, 0xab, 0x16, 0x39, 0x65, 0xf2, 0x8a, 0x1f}}
	return a, nil
}

var _migrations202106080PendingKycStatusSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\xcd\x31\x0a\xc2\x30\x14\x06\xe0\xfd\x9d\xe2\xdf\xa5\x5e\xa0\x53\x35\xdd\xa2\x95\xd2\xce\x21\xc6\x50\x83\xe6\x25\x98\x17\x8a\x9e\x5e\x70\x12\x9c\x1c\xbf\xe9\x6b\x1a\x6c\x62\x58\x1e\x56\x3c\xe6\x4c\xd4\xe9\xa9\x1f\x31\x75\x3b\xdd\x23\xd7\xf3\x3d\xb8\xad\x75\x2e\x55\x96\x62\x6e\x4f\x67\x8a\x58\xa9\x85\x00\xa0\x53\x0a\xfb\x41\xcf\x87\x23\xb2\xe7\x4b\xe0\xc5\x58\x81\x84\xe8\x8b\xd8\x98\xb1\x06\xb9\x7e\x88\x57\x62\xdf\x12\x7d\x5f\x2a\xad\xfc\xd7\xa6\xc6\xe1\xf4\xdb\xb5\xf4\x0e\x00\x00\xff\xff\x0b\x35\xb1\x8a\xc1\x00\x00\x00")

func migrations202106080PendingKycStatusSqlBytes() ([]byte, error) {
	return bindataRead(
		_migrations202106080PendingKycStatusSql,
		"migrations/2021-06-08.0.pending-kyc-status.sql",
	)
}

func migrations202106080PendingKycStatusSql() (*asset, error) {
	bytes, err := migrations202106080PendingKycStatusSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/2021-06-08.0.pending-kyc-status.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x10, 0x1c, 0x6f, 0xa9, 0x5e, 0x89, 0xfa, 0x5b, 0x1f, 0x1e, 0xf2, 0xc6, 0xe0, 0xeb, 0x6f, 0xe5, 0xa5, 0x63, 0x50, 0x6b, 0xd5, 0xdb, 0x54, 0xac, 0xc2, 0x1, 0x82, 0x27, 0xc4, 0x70, 0xcf, 0x9c}}
	return a, nil
}

var _migrations202306060FxRatesSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\xd0\x41\x4b\xc3\x40\x10\x05\xe0\xfb\xfc\x8a\x77\x6c\xb1\x11\x3c\xf7\x14\xcd\x0a\xc5\x35\x29\xdb\x04\xe9\x29\xac\xc9\xd6\x2e\x24\xd9\xb0\x33\xa1\xc5\x5f\x2f\x8d\xa0\xa5\xf4\x36\xc3\x07\xef\xc1\x4b\x12\x3c\xf4\xfe\x2b\x5a\x71\xa8\x46\xa2\x17\xa3\xd2\x52\xa1\x4c\x9f\xb5\xc2\x38\x7d\x76\xbe\x79\x3c\x9c\xeb\x8b\x33\x16\x04\x00\xbe\xc5\x4e\x99\x4d\xaa\xb1\x35\x9b\xf7\xd4\xec\xf1\xa6\xf6\xab\x99\x2c\xb3\x93\xba\x09\xad\x83\xb8\xb3\x20\x2f\x4a\xe4\x95\xd6\xd7\xea\x99\x27\x17\xef\xf9\xc4\xed\x5c\x74\x63\x40\x92\xc0\xf6\x61\x1a\x04\xe1\xf0\x9b\x02\x39\x5a\x41\x13\x62\x74\x3c\x86\xa1\x65\x48\xc0\x13\xaa\x1d\xb2\xd0\x75\x36\xce\x79\x97\xac\x5a\x7c\xef\x58\x6c\x3f\xe2\xff\x3a\x79\x39\xce\x2f\xbe\xc3\xe0\xfe\x9a\x90\xa9\xd7\xb4\xd2\x25\xf2\xe2\x63\xb1\x5c\xd1\x72\x4d\x74\xbd\x4f\x16\x4e\x03\x51\x66\x8a\xed\xfd\x7d\xd6\x44\x3f\x01\x00\x00\xff\xff\x29\x95\xea\x8e\x4f\x01\x00\x00")

func migrations202306060FxRatesSqlBytes() ([]byte, error) {
	return bindataRead(
		_migrations202306060FxRatesSql,
		"migrations/2023-06-06.0.fx-rates.sql",
	)
}

func migrations202306060FxRatesSql() (*asset, error) {
	bytes, err := migrations202306060FxRatesSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migrations/2023-06-06.0.fx-rates.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xa, 0xfe, 0x81, 0xe5, 0xfb, 0xb5, 0x30, 0x41, 0xab, 0x40, 0xe, 0x53, 0x77, 0x2f, 0x5f, 0xd, 0xcd, 0xa8, 0xb9, 0x32, 0x51, 0x8a, 0x9c, 0x15, 0x3c, 0xcd, 0x83, 0x3b, 0x48, 0xe8, 0x37, 0x54}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"migrations/2021-05-05.0.initial.sql":             migrations202105050InitialSql,
	"migrations/2021-05-18.0.accounts-kyc-status.sql": migrations202105180AccountsKycStatusSql,
	"migrations/2021-06-08.0.pending-kyc-status.sql":  migrations202106080PendingKycStatusSql,
	"migrations/2023-06-06.0.fx-rates.sql":            migrations202306060FxRatesSql,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"migrations": {nil, map[string]*bintree{
		"2021-05-05.0.initial.sql":             {migrations202105050InitialSql, map[string]*bintree{}},
		"2021-05-18.0.accounts-kyc-status.sql": {migrations202105180AccountsKycStatusSql, map[string]*bintree{}},
		"2021-06-08.0.pending-kyc-status.sql":  {migrations202106080PendingKycStatusSql, map[string]*bintree{}},
		"2023-06-06.0.fx-rates.sql":            {migrations202306060FxRatesSql, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
