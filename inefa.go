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
import "time"
import "os"

type Facade interface{
	Lookup(id *uuid.UUID,name string) (*uuid.UUID,error)
	Chtimes(id *uuid.UUID,atime time.Time, mtime time.Time) error
	Truncate(id *uuid.UUID,size int64) error
	WriteAt(id *uuid.UUID, b []byte, off int64) (int,error)
	Readdirnames(id *uuid.UUID) ([]string,error)
}

type Statbuf struct {
	Size int64
	ModTime time.Time
	IsDir bool
	IsRegular bool
}
func (s *Statbuf) FromFileInfo(i os.FileInfo) {
	s.Size      = i.Size()
	s.IsDir     = i.IsDir()
	s.IsRegular = i.Mode().IsRegular()
	s.ModTime   = i.ModTime()
}

type Facade2 interface{
	Facade
	HL_Mkdir (id *uuid.UUID,name string) (*uuid.UUID,error)
	HL_Mkfile(id *uuid.UUID,name string) (*uuid.UUID,error)
	HL_Stat  (id *uuid.UUID, sb *Statbuf) error
	HL_Delete(id *uuid.UUID,name string) error
	HL_ReadAt(id *uuid.UUID, b []byte, off int64) ([]byte,error)
	HL_Movelink(oid *uuid.UUID, oname string, nid *uuid.UUID, nname string) error
	
	// RPC-Friendly version of HL_ReadAt
	HL_ReadAt2(id *uuid.UUID, size int, off int64) ([]byte,error)
}

type LL_Facade interface{
	Facade
	Stat(id *uuid.UUID) (os.FileInfo, error)
	Mkfile(id *uuid.UUID) error
	Mkdir(id *uuid.UUID) error
	PutDirent(id *uuid.UUID,name string, child *uuid.UUID) error
	DelDirent(id *uuid.UUID,name string) error
	DelDirentFull(id *uuid.UUID,name string) error
	ReadAt(id *uuid.UUID, b []byte, off int64) (int,error)
}
type HL_Wrap struct{
	LL_Facade
}

func (h *HL_Wrap) HL_Mkdir (id *uuid.UUID,name string) (*uuid.UUID,error) {
	nid,e := uuid.NewV4()
	if e!=nil { return nil,e }
	e = h.Mkdir(nid)
	if e!=nil { return nil,e }
	e = h.PutDirent(id,name,nid)
	if e!=nil { return nil,e }
	return nid,nil
}
func (h *HL_Wrap) HL_Mkfile(id *uuid.UUID,name string) (*uuid.UUID,error) {
	nid,e := uuid.NewV4()
	if e!=nil { return nil,e }
	e = h.Mkfile(nid)
	if e!=nil { return nil,e }
	e = h.PutDirent(id,name,nid)
	if e!=nil { return nil,e }
	return nid,nil
}
func (h *HL_Wrap) HL_Stat(id *uuid.UUID, sb *Statbuf) error {
	s,e := h.Stat(id)
	if e!=nil { return e }
	sb.FromFileInfo(s)
	return nil
}
func (h *HL_Wrap) HL_Delete(id *uuid.UUID,name string) error {
	return h.DelDirentFull(id,name)
}
func (h *HL_Wrap) HL_ReadAt(id *uuid.UUID, b []byte, off int64) ([]byte,error) {
	n,e := h.ReadAt(id,b,off)
	return b[:n],e
}
func (h *HL_Wrap) HL_ReadAt2(id *uuid.UUID, size int, off int64) ([]byte,error) {
	b := make([]byte,size)
	n,e := h.ReadAt(id,b,off)
	return b[:n],e
}
func (h *HL_Wrap) HL_Movelink(oid *uuid.UUID, oname string, nid *uuid.UUID, nname string) error {
	id,e := h.Lookup(oid,oname)
	if e!=nil { return e }
	e = h.PutDirent(nid,nname,id)
	if e!=nil { return e }
	e = h.DelDirent(oid,oname)
	if e!=nil {
		h.DelDirent(nid,nname)
	}
	return e
}

