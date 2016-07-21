// Code generated by go-bindata.
// sources:
// templates/migration-main.go.tmpl
// templates/migration.go.tmpl
// templates/migration.sql.tmpl
// DO NOT EDIT!

package goose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
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

var _templatesMigrationMainGoTmpl = []byte(`package main

import (
	"log"
	"bytes"
	"encoding/gob"

	_ "{{.Import}}"
	"github.com/steinbacher/goose"
)

func main() {

	var conf goose.DBConf
	buf := bytes.NewBuffer({{ .Conf }})
	if err := gob.NewDecoder(buf).Decode(&conf); err != nil {
		log.Fatal("gob.Decode - ", err)
	}

	db, err := goose.OpenDBFromDBConf(&conf)
	if err != nil {
		log.Fatal("failed to open DB:", err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		log.Fatal("db.Begin:", err)
	}

	{{ .Func }}(txn, db)

	err = goose.FinalizeMigration(&conf, txn, goose.{{ .Direction }}, {{ .Version }})
	if err != nil {
		log.Fatal("Commit() failed:", err)
	}
}
{{/* vim: set ft=go.gotexttmpl: */}}
`)

func templatesMigrationMainGoTmplBytes() ([]byte, error) {
	return _templatesMigrationMainGoTmpl, nil
}

func templatesMigrationMainGoTmpl() (*asset, error) {
	bytes, err := templatesMigrationMainGoTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/migration-main.go.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesMigrationGoTmpl = []byte(`package main

import (
	"database/sql"
)

// Up is executed when this migration is applied
func Up_{{ . }}(txn *sql.Tx, db *sql.DB) {

}

// Down is executed when this migration is rolled back
func Down_{{ . }}(txn *sql.Tx, db *sql.DB) {

}
{{/* vim: set ft=go.gotexttmpl: */}}
`)

func templatesMigrationGoTmplBytes() ([]byte, error) {
	return _templatesMigrationGoTmpl, nil
}

func templatesMigrationGoTmpl() (*asset, error) {
	bytes, err := templatesMigrationGoTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/migration.go.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesMigrationSqlTmpl = []byte(`-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back


`)

func templatesMigrationSqlTmplBytes() ([]byte, error) {
	return _templatesMigrationSqlTmpl, nil
}

func templatesMigrationSqlTmpl() (*asset, error) {
	bytes, err := templatesMigrationSqlTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/migration.sql.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
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

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
	"templates/migration-main.go.tmpl": templatesMigrationMainGoTmpl,
	"templates/migration.go.tmpl":      templatesMigrationGoTmpl,
	"templates/migration.sql.tmpl":     templatesMigrationSqlTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
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
	"templates": &bintree{nil, map[string]*bintree{
		"migration-main.go.tmpl": &bintree{templatesMigrationMainGoTmpl, map[string]*bintree{}},
		"migration.go.tmpl":      &bintree{templatesMigrationGoTmpl, map[string]*bintree{}},
		"migration.sql.tmpl":     &bintree{templatesMigrationSqlTmpl, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
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
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
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
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
