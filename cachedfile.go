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
import "github.com/hashicorp/golang-lru"
import "os"

type CachedFileSystem struct {
	*FileSystem
	Cache *lru.Cache
}
func cacheEvict(key interface{}, value interface{}) {
	value.(*os.File).Close()
}
func (fs *CachedFileSystem) Init(f2 *FileSystem,size int) *CachedFileSystem{
	if f2!=nil { fs.FileSystem = f2 }
	if size<0 { panic("Size < 0") }
	fs.Cache,_ = lru.NewWithEvict(size,cacheEvict)
	return fs
}
func (fs *CachedFileSystem) Mkfile(id *uuid.UUID) error {
	f,e := fs.Open(id,os.O_CREATE|os.O_RDWR)
	if e!=nil { return e }
	f.Close()
	return e
}
func (fs *CachedFileSystem) getFile(id *uuid.UUID) (*os.File,error) {
	fn,ids := fs.extrude2(id)
	f,ok := fs.Cache.Get(ids)
	if ok { return f.(*os.File),nil }
	f,e := os.OpenFile(fn,os.O_RDWR,0600)
	if e!=nil { return nil,e }
	fs.Cache.Add(ids,f)
	return f.(*os.File),nil
}
func (fs *CachedFileSystem) Stat(id *uuid.UUID) (os.FileInfo, error) {
	fn,ids := fs.extrude2(id)
	f,ok := fs.Cache.Get(ids)
	if ok {
		return f.(*os.File).Stat()
	}else{
		return os.Stat(fn)
	}
}
func (fs *CachedFileSystem) Truncate(id *uuid.UUID,size int64) (error) {
	fn,ids := fs.extrude2(id)
	f,ok := fs.Cache.Get(ids)
	if ok {
		return f.(*os.File).Truncate(size)
	}else{
		return os.Truncate(fn,size)
	}
}
func (fs *CachedFileSystem) ReadAt(id *uuid.UUID, b []byte, off int64) (int,error) {
	f,e := fs.getFile(id)
	if e!=nil { return 0,e }
	return f.ReadAt(b,off)
}
func (fs *CachedFileSystem) WriteAt(id *uuid.UUID, b []byte, off int64) (int,error) {
	f,e := fs.getFile(id)
	if e!=nil { return 0,e }
	return f.WriteAt(b,off)
}
func (fs *CachedFileSystem) Readdirnames(id *uuid.UUID) ([]string,error) {
	fn := fs.extrude(id)
	f,e := os.Open(fn)
	if e!=nil { return nil,e }
	defer f.Close()
	return f.Readdirnames(0)
}

