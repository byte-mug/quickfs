/*
MIT License

Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package quickfs

import "github.com/nu7hatch/gouuid"
import "os"
import "path/filepath"
import "time"

type FileSystem struct {
	Prefix string
}
func (fs *FileSystem) extrude(id *uuid.UUID) string {
	r,_ := fs.extrude2(id)
	return r
}
func (fs *FileSystem) extrude2(id *uuid.UUID) (string,string) {
	ids := id.String()
	return fs.Prefix+id.String(),ids
}
func (fs *FileSystem) deextrude(s string) (*uuid.UUID,error){
	s = filepath.Base(s)
	return uuid.ParseHex(s)
}
func (fs *FileSystem) Open(id *uuid.UUID,flag int) (*os.File, error) {
	return os.OpenFile(fs.extrude(id),flag,0600)
}
func (fs *FileSystem) Mkdir(id *uuid.UUID) error {
	return os.Mkdir(fs.extrude(id),0700)
}
func (fs *FileSystem) Lookup(id *uuid.UUID,name string) (*uuid.UUID,error) {
	s,e := os.Readlink(fs.extrude(id)+"/"+name)
	if e!=nil { return nil,e }
	return fs.deextrude(s)
}
func (fs *FileSystem) PutDirent(id *uuid.UUID,name string, child *uuid.UUID) error {
	return os.Symlink(fs.extrude(child),fs.extrude(id)+"/"+name)
}
func (fs *FileSystem) DelDirent(id *uuid.UUID,name string) error {
	return os.Remove(fs.extrude(id)+"/"+name)
}
func (fs *FileSystem) DelDirentFull(id *uuid.UUID,name string) error {
	cld,err := os.Readlink(fs.extrude(id)+"/"+name)
	e := os.Remove(fs.extrude(id)+"/"+name)
	if e!=nil { return e }
	if err==nil {
		e = os.Remove(cld)
	}
	return e
}
func (fs *FileSystem) Stat(id *uuid.UUID) (os.FileInfo, error) {
	return os.Stat(fs.extrude(id))
}
func (fs *FileSystem) Chtimes(id *uuid.UUID,atime time.Time, mtime time.Time) error {
	return os.Chtimes(fs.extrude(id),atime,mtime)
}

