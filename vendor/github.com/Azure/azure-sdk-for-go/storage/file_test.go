package storage

import (
	"bytes"
	"io"
	"math/rand"
	"strconv"

	chk "gopkg.in/check.v1"
)

type StorageFileSuite struct{}

var _ = chk.Suite(&StorageFileSuite{})

func getFileClient(c *chk.C) FileServiceClient {
	return getBasicClient(c).GetFileService()
}

func (s *StorageFileSuite) Test_pathSegments(c *chk.C) {
	c.Assert(ToPathSegment("foo"), chk.Equals, "/foo")
	c.Assert(ToPathSegment("foo", "bar"), chk.Equals, "/foo/bar")
	c.Assert(ToPathSegment("foo", "bar", "baz"), chk.Equals, "/foo/bar/baz")
}

func (s *StorageFileSuite) TestGetURL(c *chk.C) {
	api, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	cli := api.GetFileService()

	c.Assert(cli.GetShareURL("share"), chk.Equals, "https://foo.file.core.windows.net/share")
	c.Assert(cli.GetDirectoryURL("share/dir"), chk.Equals, "https://foo.file.core.windows.net/share/dir")
}

func (s *StorageFileSuite) TestCreateShareDeleteShare(c *chk.C) {
	cli := getFileClient(c)
	name := randShare()
	c.Assert(cli.CreateShare(name, nil), chk.IsNil)
	c.Assert(cli.DeleteShare(name), chk.IsNil)
}

func (s *StorageFileSuite) TestCreateShareIfNotExists(c *chk.C) {
	cli := getFileClient(c)
	name := randShare()
	defer cli.DeleteShare(name)

	// First create
	ok, err := cli.CreateShareIfNotExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)

	// Second create, should not give errors
	ok, err = cli.CreateShareIfNotExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)
}

func (s *StorageFileSuite) TestDeleteShareIfNotExists(c *chk.C) {
	cli := getFileClient(c)
	name := randShare()

	// delete non-existing share
	ok, err := cli.DeleteShareIfExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	c.Assert(cli.CreateShare(name, nil), chk.IsNil)

	// delete existing share
	ok, err = cli.DeleteShareIfExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *StorageFileSuite) Test_checkForStorageEmulator(c *chk.C) {
	f := getEmulatorClient(c).GetFileService()
	err := f.checkForStorageEmulator()
	c.Assert(err, chk.NotNil)
}

func (s *StorageFileSuite) TestListShares(c *chk.C) {
	cli := getFileClient(c)
	c.Assert(deleteTestShares(cli), chk.IsNil)

	name := randShare()

	c.Assert(cli.CreateShare(name, nil), chk.IsNil)
	defer cli.DeleteShare(name)

	resp, err := cli.ListShares(ListSharesParameters{
		MaxResults: 5,
		Prefix:     testSharePrefix})
	c.Assert(err, chk.IsNil)

	c.Check(len(resp.Shares), chk.Equals, 1)
	c.Check(resp.Shares[0].Name, chk.Equals, name)

}

func (s *StorageFileSuite) TestShareExists(c *chk.C) {
	cli := getFileClient(c)
	name := randShare()

	ok, err := cli.ShareExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	c.Assert(cli.CreateShare(name, nil), chk.IsNil)
	defer cli.DeleteShare(name)

	ok, err = cli.ShareExists(name)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *StorageFileSuite) TestGetAndSetShareProperties(c *chk.C) {
	name := randShare()
	quota := rand.Intn(5120)

	cli := getFileClient(c)
	c.Assert(cli.CreateShare(name, nil), chk.IsNil)
	defer cli.DeleteShare(name)

	err := cli.SetShareProperties(name, ShareHeaders{Quota: strconv.Itoa(quota)})
	c.Assert(err, chk.IsNil)

	props, err := cli.GetShareProperties(name)
	c.Assert(err, chk.IsNil)

	c.Assert(props.Quota, chk.Equals, strconv.Itoa(quota))
}

func (s *StorageFileSuite) TestGetAndSetShareMetadata(c *chk.C) {
	cli := getFileClient(c)
	share1 := randShare()

	c.Assert(cli.CreateShare(share1, nil), chk.IsNil)
	defer cli.DeleteShare(share1)

	m, err := cli.GetShareMetadata(share1)
	c.Assert(err, chk.IsNil)
	c.Assert(m, chk.Not(chk.Equals), nil)
	c.Assert(len(m), chk.Equals, 0)

	share2 := randShare()
	mCreate := map[string]string{
		"create": "data",
	}
	c.Assert(cli.CreateShare(share2, mCreate), chk.IsNil)
	defer cli.DeleteShare(share2)

	m, err = cli.GetShareMetadata(share2)
	c.Assert(err, chk.IsNil)
	c.Assert(m, chk.Not(chk.Equals), nil)
	c.Assert(len(m), chk.Equals, 1)

	mPut := map[string]string{
		"foo":     "bar",
		"bar_baz": "waz qux",
	}

	err = cli.SetShareMetadata(share2, mPut)
	c.Assert(err, chk.IsNil)

	m, err = cli.GetShareMetadata(share2)
	c.Assert(err, chk.IsNil)
	c.Check(m, chk.DeepEquals, mPut)

	// Case munging

	mPutUpper := map[string]string{
		"Foo":     "different bar",
		"bar_BAZ": "different waz qux",
	}
	mExpectLower := map[string]string{
		"foo":     "different bar",
		"bar_baz": "different waz qux",
	}

	err = cli.SetShareMetadata(share2, mPutUpper)
	c.Assert(err, chk.IsNil)

	m, err = cli.GetShareMetadata(share2)
	c.Assert(err, chk.IsNil)
	c.Check(m, chk.DeepEquals, mExpectLower)
}

func (s *StorageFileSuite) TestListDirsAndFiles(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	// list contents, should be empty
	resp, err := cli.ListDirsAndFiles(share, ListDirsAndFilesParameters{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Directories, chk.IsNil)
	c.Assert(resp.Files, chk.IsNil)

	// create a directory and a file
	dir := "SomeDirectory"
	file := "foo.file"
	c.Assert(cli.CreateDirectory(ToPathSegment(share, dir), nil), chk.IsNil)
	c.Assert(cli.CreateFile(ToPathSegment(share, file), 512, nil), chk.IsNil)

	// list contents
	resp, err = cli.ListDirsAndFiles(share, ListDirsAndFilesParameters{})
	c.Assert(err, chk.IsNil)
	c.Assert(len(resp.Directories), chk.Equals, 1)
	c.Assert(len(resp.Files), chk.Equals, 1)
	c.Assert(resp.Directories[0].Name, chk.Equals, dir)
	c.Assert(resp.Files[0].Name, chk.Equals, file)
}

func (s *StorageFileSuite) TestCreateDirectory(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	// directory shouldn't exist
	dir := ToPathSegment(share, "SomeDirectory")
	exists, err := cli.DirectoryExists(dir)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create directory
	exists, err = cli.CreateDirectoryIfNotExists(dir)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	// try to create again, should fail
	c.Assert(cli.CreateDirectory(dir, nil), chk.NotNil)
	exists, err = cli.CreateDirectoryIfNotExists(dir)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// get properties
	var props *DirectoryProperties
	props, err = cli.GetDirectoryProperties(dir)
	c.Assert(props.Etag, chk.Not(chk.Equals), "")
	c.Assert(props.LastModified, chk.Not(chk.Equals), "")

	// delete directory and verify
	c.Assert(cli.DeleteDirectory(dir), chk.IsNil)
	exists, err = cli.DirectoryExists(dir)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)
}

func (s *StorageFileSuite) TestCreateFile(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	// create directory structure
	dir1 := ToPathSegment(share, "one")
	c.Assert(cli.CreateDirectory(dir1, nil), chk.IsNil)
	dir2 := ToPathSegment(dir1, "two")
	c.Assert(cli.CreateDirectory(dir2, nil), chk.IsNil)

	// verify file doesn't exist
	file := ToPathSegment(dir2, "some.file")
	exists, err := cli.FileExists(file)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create file
	c.Assert(cli.CreateFile(file, 1024, nil), chk.IsNil)
	exists, err = cli.FileExists(file)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	// delete file and verify
	c.Assert(cli.DeleteFile(file), chk.IsNil)
	exists, err = cli.FileExists(file)
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)
}

func (s *StorageFileSuite) TestGetFile(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	// create file
	const size = uint64(1024)
	file := ToPathSegment(share, "some.file")
	c.Assert(cli.CreateFile(file, size, nil), chk.IsNil)

	// fill file with some data
	c.Assert(cli.PutRange(file, newByteStream(size), FileRange{End: size - 1}), chk.IsNil)

	// set some metadata
	md := map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	c.Assert(cli.SetFileMetadata(file, md), chk.IsNil)

	// retrieve full file content and verify
	stream, err := cli.GetFile(file, nil)
	c.Assert(err, chk.IsNil)
	defer stream.Body.Close()
	var b1 [size]byte
	count, _ := stream.Body.Read(b1[:])
	c.Assert(count, chk.Equals, int(size))
	var c1 [size]byte
	newByteStream(size).Read(c1[:])
	c.Assert(b1, chk.DeepEquals, c1)
	c.Assert(stream.Properties.ContentLength, chk.Equals, size)
	c.Assert(stream.Metadata, chk.DeepEquals, md)

	// retrieve partial file content and verify
	stream, err = cli.GetFile(file, &FileRange{Start: size / 2, End: size - 1})
	c.Assert(err, chk.IsNil)
	defer stream.Body.Close()
	var b2 [size / 2]byte
	count, _ = stream.Body.Read(b2[:])
	c.Assert(count, chk.Equals, int(size)/2)
	var c2 [size / 2]byte
	newByteStream(size / 2).Read(c2[:])
	c.Assert(b2, chk.DeepEquals, c2)
}

func (s *StorageFileSuite) TestFileRanges(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	// create file
	fileSize := uint64(4096)
	file := ToPathSegment(share, "test.dat")
	c.Assert(cli.CreateFile(file, fileSize, nil), chk.IsNil)

	// verify there are no valid ranges
	ranges, err := cli.ListFileRanges(file, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.ContentLength, chk.Equals, fileSize)
	c.Assert(ranges.FileRanges, chk.IsNil)

	// fill entire range and validate
	c.Assert(cli.PutRange(file, newByteStream(fileSize), FileRange{End: fileSize - 1}), chk.IsNil)
	ranges, err = cli.ListFileRanges(file, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(len(ranges.FileRanges), chk.Equals, 1)
	c.Assert((ranges.FileRanges[0].End-ranges.FileRanges[0].Start)+1, chk.Equals, fileSize)

	// clear entire range and validate
	c.Assert(cli.ClearRange(file, FileRange{End: fileSize - 1}), chk.IsNil)
	ranges, err = cli.ListFileRanges(file, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.IsNil)

	// put partial ranges on 512 byte aligned boundaries
	putRanges := []FileRange{
		{End: 511},
		{Start: 1024, End: 1535},
		{Start: 2048, End: 2559},
		{Start: 3072, End: 3583},
	}

	for _, r := range putRanges {
		err = cli.PutRange(file, newByteStream(512), r)
		c.Assert(err, chk.IsNil)
	}

	// validate all ranges
	ranges, err = cli.ListFileRanges(file, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.DeepEquals, putRanges)

	// validate sub-ranges
	ranges, err = cli.ListFileRanges(file, &FileRange{Start: 1000, End: 3000})
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.DeepEquals, putRanges[1:3])

	// clear partial range and validate
	c.Assert(cli.ClearRange(file, putRanges[0]), chk.IsNil)
	c.Assert(cli.ClearRange(file, putRanges[2]), chk.IsNil)
	ranges, err = cli.ListFileRanges(file, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(len(ranges.FileRanges), chk.Equals, 2)
	c.Assert(ranges.FileRanges[0], chk.DeepEquals, putRanges[1])
	c.Assert(ranges.FileRanges[1], chk.DeepEquals, putRanges[3])
}

func (s *StorageFileSuite) TestFileProperties(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	fileSize := uint64(512)
	file := ToPathSegment(share, "test.dat")
	c.Assert(cli.CreateFile(file, fileSize, nil), chk.IsNil)

	// get initial set of properties
	props, err := cli.GetFileProperties(file)
	c.Assert(err, chk.IsNil)
	c.Assert(props.ContentLength, chk.Equals, fileSize)

	// set some file properties
	cc := "cachecontrol"
	ct := "mytype"
	enc := "noencoding"
	lang := "neutral"
	disp := "friendly"
	props.CacheControl = cc
	props.ContentType = ct
	props.Disposition = disp
	props.Encoding = enc
	props.Language = lang
	c.Assert(cli.SetFileProperties(file, *props), chk.IsNil)

	// retrieve and verify
	props, err = cli.GetFileProperties(file)
	c.Assert(err, chk.IsNil)
	c.Assert(props.CacheControl, chk.Equals, cc)
	c.Assert(props.ContentType, chk.Equals, ct)
	c.Assert(props.Disposition, chk.Equals, disp)
	c.Assert(props.Encoding, chk.Equals, enc)
	c.Assert(props.Language, chk.Equals, lang)
}

func (s *StorageFileSuite) TestDirectoryMetadata(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	dir1 := ToPathSegment(share, "testdir1")
	c.Assert(cli.CreateDirectory(dir1, nil), chk.IsNil)

	// get metadata, shouldn't be any
	md, err := cli.GetDirectoryMetadata(dir1)
	c.Assert(err, chk.IsNil)
	c.Assert(md, chk.HasLen, 0)

	mCreate := map[string]string{
		"create": "data",
	}
	dir2 := ToPathSegment(share, "testdir2")
	c.Assert(cli.CreateDirectory(dir2, mCreate), chk.IsNil)

	// get metadata
	md, err = cli.GetDirectoryMetadata(dir2)
	c.Assert(err, chk.IsNil)
	c.Assert(md, chk.HasLen, 1)

	// set some custom metadata
	md = map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	c.Assert(cli.SetDirectoryMetadata(dir2, md), chk.IsNil)

	// retrieve and verify
	var mdRes map[string]string
	mdRes, err = cli.GetDirectoryMetadata(dir2)
	c.Assert(err, chk.IsNil)
	c.Assert(mdRes, chk.DeepEquals, md)
}

func (s *StorageFileSuite) TestFileMetadata(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := randShare()

	c.Assert(cli.CreateShare(share, nil), chk.IsNil)
	defer cli.DeleteShare(share)

	fileSize := uint64(512)
	file1 := ToPathSegment(share, "test1.dat")
	c.Assert(cli.CreateFile(file1, fileSize, nil), chk.IsNil)

	// get metadata, shouldn't be any
	md, err := cli.GetFileMetadata(file1)
	c.Assert(err, chk.IsNil)
	c.Assert(md, chk.HasLen, 0)

	mCreate := map[string]string{
		"create": "data",
	}
	file2 := ToPathSegment(share, "test2.dat")
	c.Assert(cli.CreateFile(file2, fileSize, mCreate), chk.IsNil)

	// get metadata
	md, err = cli.GetFileMetadata(file2)
	c.Assert(err, chk.IsNil)
	c.Assert(md, chk.HasLen, 1)

	// set some custom metadata
	md = map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	c.Assert(cli.SetFileMetadata(file2, md), chk.IsNil)

	// retrieve and verify
	var mdRes map[string]string
	mdRes, err = cli.GetFileMetadata(file2)
	c.Assert(err, chk.IsNil)
	c.Assert(mdRes, chk.DeepEquals, md)
}

func deleteTestShares(cli FileServiceClient) error {
	for {
		resp, err := cli.ListShares(ListSharesParameters{Prefix: testSharePrefix})
		if err != nil {
			return err
		}
		if len(resp.Shares) == 0 {
			break
		}
		for _, c := range resp.Shares {
			err = cli.DeleteShare(c.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const testSharePrefix = "zzzzztest"

func randShare() string {
	return testSharePrefix + randString(32-len(testSharePrefix))
}

func newByteStream(count uint64) io.Reader {
	b := make([]uint8, count)
	for i := uint64(0); i < count; i++ {
		b[i] = 0xff
	}
	return bytes.NewReader(b)
}
