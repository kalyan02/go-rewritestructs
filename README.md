# go-rewritestructs

This tool rewrites the individual field in structure declarations of a go source file to be pointer types as defined in `types.json`

## How does it work?

- Tool parses the syntax tree generated using the `go/parser` `go/ast` packages
- Walks the tree using `github.com/fatih/astrewrite`
- Syntax tree nodes are rewritten to make these changes
  - variables of desired types (from types.json) are made into pointer types
  - arrays of desired types are pointer types
  - value type in map type are rewritten (but not keys)
- Source code is re-generated again using  `go/format` package.

## Limitations

- Member methods are not rewritten to use pointers
- Usage is not rewritten. 

## Usage

    ./go-declstructptrs -dir test-source -types test-source/types.json 

    -write        : overwrite in place otherwise print
    -dir   <dir>  : directory of files
    -file  <file> : single file


## License

BSD
