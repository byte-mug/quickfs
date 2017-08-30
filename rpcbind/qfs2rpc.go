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


// RPC Interface for Quickfs
package rpcbind

import "github.com/byte-mug/quickfs"
import "net/rpc"
import "github.com/nu7hatch/gouuid"
import "errors"
import "time"

/*
func slaughter(id *uuid.UUID) [16]byte {
	raw := *id
	return [16]byte(raw)
}
*/
func slaughter(id *uuid.UUID) []byte {
	if id==nil { return nil }
	return id[:]
}
func joine(es...error) error {
	for _,e := range es {
		if e!=nil { return e }
	}
	return nil
}
func join2(a,b error) error {
	if a!=nil { return a }
	return b
}
func join3(a,b,c error) error {
	if a!=nil { return a }
	if b!=nil { return b }
	return c
}

type Errcon struct{
	Msg string
	Bad bool
}
func (e *Errcon) From(r error) error {
	if r!=nil {
		*e = Errcon{r.Error(),true}
	}else{
		*e = Errcon{"",false}
	}
	return nil
}
func (e *Errcon) To() error {
	if !e.Bad { return nil }
	return errors.New(e.Msg)
}

// Wraps a RPC client into a QuickFS facade
func FacadeFrom(c *rpc.Client) quickfs.Facade2 {
	return &QuickfsClient{c}
}

// Adds a QuickFS facade to an RPC service. Only one per rpc.Server can be added.
func FacadeTo(f quickfs.Facade2,s *rpc.Server) error {
	return s.Register(&QuickfsFacade{f})
}

type QuickfsFacade struct{
	Facade quickfs.Facade2
}
type QuickfsClient struct{
	Client *rpc.Client
}

type QLookup struct{
	Id []byte
	Name string
}
type ALookup struct{
	Id []byte
	Err Errcon
}
func (f *QuickfsFacade) Lookup(q *QLookup, a *ALookup) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	id,e = f.Facade.Lookup(id,q.Name)
	a.Id = slaughter(id)
	return a.Err.From(e)
}
func (c *QuickfsClient) Lookup(id *uuid.UUID,name string) (*uuid.UUID,error) {
	var q QLookup
	var a ALookup
	q.Id = slaughter(id)
	q.Name = name
	e3 := c.Client.Call("QuickfsFacade.Lookup",q,&a)
	nid,e2 := uuid.Parse(a.Id)
	e1 := a.Err.To()
	return nid,join3(e1,e2,e3)
}
type QChtimes struct{
	Id []byte
	A,M time.Time
}
func (f *QuickfsFacade) Chtimes(q *QChtimes, a *Errcon) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.From(e) }
	e = f.Facade.Chtimes(id,q.A,q.M)
	return a.From(e)
}
func (c *QuickfsClient) Chtimes(id *uuid.UUID,atime time.Time, mtime time.Time) error {
	var q QChtimes
	var a Errcon
	q.Id = slaughter(id)
	q.A = atime
	q.M = mtime
	e2 := c.Client.Call("QuickfsFacade.Chtimes",q,&a)
	e1 := a.To()
	return join2(e1,e2)
}
type QTruncate struct{
	Id []byte
	Size int64
}
func (f *QuickfsFacade) Truncate(q *QTruncate, a *Errcon) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.From(e) }
	e = f.Facade.Truncate(id,q.Size)
	return a.From(e)
}
func (c *QuickfsClient) Truncate(id *uuid.UUID,size int64) error {
	var q QTruncate
	var a Errcon
	q.Id = slaughter(id)
	q.Size = size
	e2 := c.Client.Call("QuickfsFacade.Truncate",q,&a)
	e1 := a.To()
	return join2(e1,e2)
}
type QWriteAt struct{
	Id []byte
	Data []byte
	Off int64
}
type AWriteAt struct{
	Size int
	Err Errcon
}
func (f *QuickfsFacade) WriteAt(q *QWriteAt, a *AWriteAt) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	i,e := f.Facade.WriteAt(id,q.Data,q.Off)
	a.Size = i
	return a.Err.From(e)
}
func (c *QuickfsClient) WriteAt(id *uuid.UUID, b []byte, off int64) (int,error) {
	var q QWriteAt
	var a AWriteAt
	q.Id = slaughter(id)
	q.Data = b
	q.Off = off
	e2 := c.Client.Call("QuickfsFacade.WriteAt",q,&a)
	e1 := a.Err.To()
	return a.Size,join2(e1,e2)
}
type QReaddir struct{
	Id []byte
}
type AReaddir struct{
	Names []string
	Err Errcon
}
func (f *QuickfsFacade) Readdir(q *QReaddir, a *AReaddir) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	names,e := f.Facade.Readdirnames(id)
	a.Names = names
	return a.Err.From(e)
}
func (c *QuickfsClient) Readdirnames(id *uuid.UUID) ([]string,error) {
	var q QReaddir
	var a AReaddir
	q.Id = slaughter(id)
	e2 := c.Client.Call("QuickfsFacade.Readdir",q,&a)
	e1 := a.Err.To()
	return a.Names,join2(e1,e2)
}

func (f *QuickfsFacade) HLMkdir(q *QLookup, a *ALookup) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	id,e = f.Facade.HL_Mkdir(id,q.Name)
	a.Id = slaughter(id)
	return a.Err.From(e)
}
func (c *QuickfsClient) HL_Mkdir(id *uuid.UUID,name string) (*uuid.UUID,error) {
	var q QLookup
	var a ALookup
	q.Id = slaughter(id)
	q.Name = name
	e3 := c.Client.Call("QuickfsFacade.HLMkdir",q,&a)
	nid,e2 := uuid.Parse(a.Id)
	e1 := a.Err.To()
	return nid,join3(e1,e2,e3)
}
func (f *QuickfsFacade) HLMkfile(q *QLookup, a *ALookup) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	id,e = f.Facade.HL_Mkfile(id,q.Name)
	a.Id = slaughter(id)
	return a.Err.From(e)
}
func (c *QuickfsClient) HL_Mkfile(id *uuid.UUID,name string) (*uuid.UUID,error) {
	var q QLookup
	var a ALookup
	q.Id = slaughter(id)
	q.Name = name
	e3 := c.Client.Call("QuickfsFacade.HLMkfile",q,&a)
	nid,e2 := uuid.Parse(a.Id)
	e1 := a.Err.To()
	return nid,join3(e1,e2,e3)
}

type QStat struct{
	Id []byte
}
type AStat struct{
	Sb quickfs.Statbuf
	Err Errcon
}

func (f *QuickfsFacade) HLStat(q *QStat, a *AStat) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	e = f.Facade.HL_Stat(id,&(a.Sb))
	return a.Err.From(e)
}
func (c *QuickfsClient) HL_Stat  (id *uuid.UUID, sb *quickfs.Statbuf) error {
	var q QStat
	var a AStat
	q.Id = slaughter(id)
	e := c.Client.Call("QuickfsFacade.HLStat",q,&a)
	if sb!=nil { *sb = a.Sb }
	return join2(a.Err.To(),e)
}

func (f *QuickfsFacade) HLDelete(q *QLookup, a *Errcon) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.From(e) }
	e = f.Facade.HL_Delete(id,q.Name)
	return a.From(e)
}
func (c *QuickfsClient) HL_Delete(id *uuid.UUID,name string) error {
	var q QLookup
	var a Errcon
	q.Id = slaughter(id)
	q.Name = name
	e2 := c.Client.Call("QuickfsFacade.HLDelete",q,&a)
	e1 := a.To()
	return join2(e1,e2)
}



type QReadAt struct{
	Id []byte
	Size int
	Off int64
}
type AReadAt struct{
	Data []byte
	Err  Errcon
}
func (f *QuickfsFacade) HLReadAt(q *QReadAt,a *AReadAt) error {
	id,e := uuid.Parse(q.Id)
	if e!=nil { return a.Err.From(e) }
	b,e := f.Facade.HL_ReadAt2(id,q.Size,q.Off)
	a.Data = b
	return a.Err.From(e)
}
func (c *QuickfsClient) HL_ReadAt(id *uuid.UUID, b []byte, off int64) ([]byte,error) {
	var q QReadAt
	var a AReadAt
	q.Id = slaughter(id)
	q.Size = len(b)
	q.Off = off
	e2 := c.Client.Call("QuickfsFacade.HLReadAt",q,&a)
	e1 := a.Err.To()
	return a.Data,join2(e1,e2)
}
func (c *QuickfsClient) HL_ReadAt2(id *uuid.UUID, size int, off int64) ([]byte,error) {
	var q QReadAt
	var a AReadAt
	q.Id = slaughter(id)
	q.Size = size
	q.Off = off
	e2 := c.Client.Call("QuickfsFacade.HLReadAt",q,&a)
	e1 := a.Err.To()
	return a.Data,join2(e1,e2)
}



type QMovelink struct {
	Oid, Nid []byte
	Oname, Nname string
}

func (f *QuickfsFacade) HLMovelink(q *QMovelink, a *Errcon) error {
	oid,e := uuid.Parse(q.Oid)
	if e!=nil { return a.From(e) }
	nid,e := uuid.Parse(q.Nid)
	if e!=nil { return a.From(e) }
	e = f.Facade.HL_Movelink(oid,q.Oname,nid,q.Nname)
	return a.From(e)
}
func (c *QuickfsClient) HL_Movelink(oid *uuid.UUID, oname string, nid *uuid.UUID, nname string) error {
	var q QMovelink
	var a Errcon
	q.Oid = slaughter(oid)
	q.Nid = slaughter(nid)
	q.Oname = oname
	q.Nname = nname
	e2 := c.Client.Call("QuickfsFacade.HLMovelink",q,&a)
	e1 := a.To()
	return join2(e1,e2)
}



