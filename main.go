package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"context"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("Expected a mountpoint arg")
	}

	mountpoint := flag.Arg(0)

	if err := mount(mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}

	filesys := &FS{}

	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	<-c.Ready

	if err := c.MountError; err != nil {
		return err
	}

	return nil
}

type FS struct {
}

var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	println("DEBUG building root")

	repos, err := GetRepos("Service")
	if err != nil {
		panic(err)
	}

	repoNames := []string{}
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}

	n := &Dir{subDirs: []Dir{
		Dir{name: "teams", files: []string{"file-test"}},
		Dir{name: "repos", subDirs: []Dir{
			Dir{name: "Service", files: repoNames},
			Dir{name: "Library"},
		}},
		Dir{name: "whatsrunningwhere"},
	},
		files: []string{"fileFOO.txt"},
	}
	return n, nil
}

type Dir struct {
	name    string
	files   []string
	subDirs []Dir
}

var _ fs.Node = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0755
	attr.Size = 512
	attr.Mtime = time.Now()
	attr.Ctime = time.Now()
	attr.Crtime = time.Now()
	return nil
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := req.Name

	/*
		if path == d.name {
			teams, err := GetTeams()
			if err != nil {
				panic(err)
			}

			files := []string{}
			for _, t := range teams {
				files = append(files, t.Name)
			}

			n := &Dir{files: files}
			return n, nil
		}*/

	for _, sd := range d.subDirs {
		if path == sd.name {
			println("found: " + path)
			n := &Dir{name: sd.name, files: sd.files, subDirs: sd.subDirs}
			return n, nil
		}
	}

	for _, f := range d.files {
		if f == path {
			println("found: " + path)
			child := &File{
				text: "",
			}
			return child, nil
		}
	}

	return nil, fuse.ENOENT
}

var _ = fs.HandleReadDirAller(&Dir{})

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var res []fuse.Dirent

	for _, sd := range d.subDirs {
		var de fuse.Dirent
		de.Name = sd.name
		res = append(res, de)
	}

	for _, d := range d.files {
		var de fuse.Dirent
		de.Name = d
		res = append(res, de)
	}

	return res, nil
}

type File struct {
	text string
}

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {

	println("File.Attr called")

	attr.Mode = 0755
	attr.Size = uint64(0)
	attr.Mtime = time.Now()
	attr.Ctime = time.Now()
	attr.Crtime = time.Now()

	return nil
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	println("File.Open called")

	if f.text == "" {
		if b, err := GetRepo("vmv"); err != nil {
			log.Println(err)
			f.text = "error"
		} else {
			f.text = b
		}
	}

	println(f.text)
	r := io.NopCloser(strings.NewReader(f.text))
	resp.Flags |= fuse.OpenNonSeekable
	return &FileHandle{r: r}, nil
}

type FileHandle struct {
	r io.ReadCloser
}

var _ fs.Handle = (*FileHandle)(nil)

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return fh.r.Close()
}

var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	println("FileHander read called")
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	resp.Data = buf[:n]

	return err
}
