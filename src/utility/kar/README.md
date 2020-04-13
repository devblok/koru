# kar - Koru Archive

kar is a package that provides a file format for use in memory mapping. Storage of read-only assets that need to be loaded into usable state quickly. Every file in the archive is individually compressed, while the archive itself is kept raw. Files are compressed with lz4, thus they decompress very quickly on the fly.

#### Properties of kar
- [x] performant with mmap
- [x] index of files is in the header, it is known prior reading exactly where the files are and how big they are
- [x] non-appendable, intended to be a read-only distributable archive
- [x] safe to use concurrently