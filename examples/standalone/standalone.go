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

package main

import _ "quickfs/rpcbind"
import "quickfs"
import "quickfs/fusebind"
import "github.com/nu7hatch/gouuid"
import "fmt"

import "github.com/hanwen/go-fuse/fuse"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "flag"
import "os"

func withSuffix(path string) string {
	if len(path)==0 { return "" }
	if path[len(path)-1]!='/' { return path+"/" }
	return path
}

func main(){
	
	// Scans the arg list and sets up flags
	debug := flag.Bool("debug", false, "print debugging messages.")
	flag.Parse()
	if flag.NArg() < 2 {
		// TODO - where to get program name?
		fmt.Println("usage: main MOUNTPOINT BACKING_STORE")
		os.Exit(2)
	}

	mountPoint := flag.Arg(0)
	backingStore := withSuffix(flag.Arg(1))
	
	
	// Make the QuickFS
	fs := &quickfs.FileSystem{backingStore}
	cfs := new(quickfs.CachedFileSystem).Init(fs,128)
	
	cfs.Mkdir(uuid.NamespaceURL)
	var facade quickfs.Facade2
	facade = &quickfs.HL_Wrap{cfs}
	
	// Make the Fuse
	
	var root nodefs.Node
	
	root = fusebind.NewOpNode(facade,uuid.NamespaceURL)
	
	conn := nodefs.NewFileSystemConnector(root, nil)
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, &fuse.MountOptions{
		Debug: *debug,
	})
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}
	fusebind.Debug = *debug
	fmt.Println("Mounted!")
	server.Serve()
}


