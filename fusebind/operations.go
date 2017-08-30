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

// Fuse-Binding for QuickFS
package fusebind

import "github.com/byte-mug/quickfs"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/hanwen/go-fuse/fuse/nodefs"

import "github.com/nu7hatch/gouuid"
import "os"
import "time"

import "fmt"

func setAttr(attr *fuse.Attr,sb *quickfs.Statbuf) {
	attr.Size = uint64(sb.Size)
	if sb.IsDir {
		attr.Mode = fuse.S_IFDIR|0777
	}else if sb.IsRegular {
		attr.Mode = fuse.S_IFREG|0666
	}
	t := sb.ModTime
	attr.SetTimes(&t,&t,&t)
}
func istrunc(flags uint32) bool {
	return (flags&uint32(os.O_TRUNC))!=0
}
func isIllegal(name string) bool {
	switch name{
	case "",".","..": return true
	}
	for _,b := range []byte(name) {
		switch b {
		case '/','\\': return true
		}
	}
	return false
}

var Debug = false
func debugln(i ...interface{}) {
	if Debug {
		fmt.Println(i...)
	}
}

type OpNode struct{
	nodefs.Node
	Facade quickfs.Facade2
	ID *uuid.UUID
}

// Wraps a QuickFS facade into a fuse nodefs.Node.
func NewOpNode(fs quickfs.Facade2,id *uuid.UUID) *OpNode {
	return &OpNode{nodefs.NewDefaultNode(),fs,id}
}
func (n *OpNode) asFile() nodefs.File {
	//return &OpFile{nodefs.NewDefaultFile(),n.Facade,n.ID}
	return nodefs.NewDefaultFile()
}
func (n *OpNode) Lookup(out *fuse.Attr, name string, context *fuse.Context) (*nodefs.Inode, fuse.Status) {
	var sb quickfs.Statbuf
	id,e := n.Facade.Lookup(n.ID,name)
	if e!=nil { return nil,fuse.ENOENT }
	if n.Facade.HL_Stat(id,&sb)!=nil { return nil,fuse.ENOENT }
	nn := &OpNode{nodefs.NewDefaultNode(),n.Facade,id}
	if out!=nil { setAttr(out,&sb) }
	return n.Inode().NewChild(name,sb.IsDir,nn),fuse.OK
}
func (n *OpNode) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	id,e := n.Facade.HL_Mkfile(n.ID,name)
	if e!=nil { return nil,fuse.EIO }
	nn := &OpNode{nodefs.NewDefaultNode(),n.Facade,id}
	return n.Inode().NewChild(name,false,nn),fuse.OK
}
func (n *OpNode) Mkdir(name string, mode uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	id,e := n.Facade.HL_Mkdir(n.ID,name)
	if e!=nil { return nil,fuse.EIO }
	nn := &OpNode{nodefs.NewDefaultNode(),n.Facade,id}
	return n.Inode().NewChild(name,true,nn),fuse.OK
}
func (n *OpNode) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	e := n.Facade.HL_Delete(n.ID,name)
	if e !=nil { return fuse.ENOENT }
	n.Inode().RmChild(name)
	return fuse.OK
}
func (n *OpNode) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	return n.Unlink(name,context)
}
func (n *OpNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, child *nodefs.Inode, code fuse.Status) {
	id,e := n.Facade.HL_Mkfile(n.ID,name)
	if e!=nil { return nil,nil,fuse.EIO }
	nn := &OpNode{nodefs.NewDefaultNode(),n.Facade,id}
	if istrunc(flags) { n.Facade.Truncate(n.ID,0) }
	return nn.asFile(),n.Inode().NewChild(name,false,nn),fuse.OK
}
func (n *OpNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	if n.Inode().IsDir() { return nil,fuse.EIO } // EISDIR
	if istrunc(flags) { n.Facade.Truncate(n.ID,0) }
	return n.asFile(),fuse.OK
}

func (n *OpNode) OpenDir(context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	s,e := n.Facade.Readdirnames(n.ID)
	if e!=nil { return nil,fuse.ENOTDIR }
	buf := make([]fuse.DirEntry,0,len(s))
	for _,name := range s {
		if isIllegal(name) { continue }
		buf = append(buf,fuse.DirEntry{Name:name})
	}
	return buf,fuse.OK
}
func (n *OpNode) Read(file nodefs.File, dest []byte, off int64, context *fuse.Context) (fuse.ReadResult, fuse.Status) {
	b,e := n.Facade.HL_ReadAt(n.ID,dest,off)
	if e!=nil && len(b)==0 { return nil,fuse.EIO }
	return fuse.ReadResultData(b),fuse.OK
}
func (n *OpNode) Write(file nodefs.File, data []byte, off int64, context *fuse.Context) (written uint32, code fuse.Status) {
	r,e := n.Facade.WriteAt(n.ID,data,off)
	if e!=nil { return uint32(r),fuse.EIO }
	return uint32(r),fuse.OK
}
func (n *OpNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	var sb quickfs.Statbuf
	if n.Facade.HL_Stat(n.ID,&sb)!=nil { return fuse.ENOENT }
	if out!=nil { setAttr(out,&sb) }
	return fuse.OK
}
func (n *OpNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	e := n.Facade.Truncate(n.ID,int64(size))
	if e!=nil { return fuse.EIO }
	return fuse.OK
}
func (n *OpNode) Utimens(file nodefs.File, atime *time.Time, mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	e := n.Facade.Chtimes(n.ID,*atime,*mtime)
	if e!=nil { return fuse.EIO }
	return fuse.OK
}
func (n *OpNode) Rename(oldName string, newParent nodefs.Node, newName string, context *fuse.Context) (code fuse.Status) {
	m,ok := newParent.(*OpNode)
	if !ok { return fuse.EIO }
	if m.Facade != n.Facade { return fuse.EIO }
	e := n.Facade.HL_Movelink(n.ID,oldName,m.ID,newName)
	if e!=nil { return fuse.EIO }
	return fuse.OK
}

/*
type OpFile struct{
	nodefs.File
	Facade quickfs.Facade2
	ID *uuid.UUID
}
*/

