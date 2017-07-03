package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/astrewrite"
)

type typesListType map[string]bool

func (t typesListType) LoadList(types []string) {
	for _, typeName := range types {
		t[typeName] = true
	}
}

func (t typesListType) Contains(typeName string) bool {
	if _, tOk := t[typeName]; tOk {
		return true
	}

	return false
}

var typesList typesListType

func main() {

	pFile := flag.String("file", "", "a file")
	pDir := flag.String("dir", "", "a directory")
	pWrite := flag.Bool("rewrite", false, "rewrite in place?")
	pTypesFile := flag.String("types", "", "json file with types to rewrite")

	flag.Parse()

	if *pDir == "" && *pFile == "" {
		fmt.Printf("Error: -dir flag is required\n")
		flag.Usage()
	}

	var files []string
	var err error

	typesList = make(typesListType)

	if *pTypesFile == "" {
		fmt.Printf("Error: -types json file is required\n")
		os.Exit(-1)
	} else {

		typesFileCon, err := ioutil.ReadFile(*pTypesFile)
		if err != nil {
			fmt.Printf("Could not read file %s: %v\n", *pTypesFile, err)
			os.Exit(-1)
		}

		var types []string
		if err := json.Unmarshal(typesFileCon, &types); err != nil {
			fmt.Printf("Could not read file %s: %v\n", *pTypesFile, err)
			os.Exit(-1)
		}

		typesList.LoadList(types)
	}

	if *pFile != "" {
		files = []string{*pFile}
	} else {
		files, err = filepath.Glob(fmt.Sprintf("%s/*.go", *pDir))
		if err != nil {
			fmt.Printf("Could not read directory %s : %v\n", *pDir, err)
			os.Exit(-1)
		}
	}

	// token store
	fset := token.NewFileSet()

	for _, file := range files {
		destfile := ""
		if *pWrite {
			destfile = file
		}
		rewriteFileDecls(fset, file, destfile)
	}

}

func rewriteFileDecls(fset *token.FileSet, inFilename string, outFilename string) {

	f, err := parser.ParseFile(fset, inFilename, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", inFilename, err)
		os.Exit(-1)
	}

	var buf bytes.Buffer

	astrewrite.Walk(f, func(n ast.Node) (ast.Node, bool) {
		x, ok := n.(*ast.StructType)
		if !ok {
			return n, true
		}

		fl := make([]*ast.Field, 0)

		// Loop through fields
		for _, fld := range x.Fields.List {

			// struct composition
			if fld.Names == nil {
				fl = append(fl, fld)
				continue
			}

			fldName := fld.Names[0].Name

			// No json tag present
			if fld.Tag == nil {
				fld.Tag = &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("`json:\"%s\"`", strings.ToLower(fldName)),
				}
				logMsgf("Rewriting %s: adding json tag", fldName)
			}

			// Is Non-Pointer type
			if idVal, idOk := fld.Type.(*ast.Ident); idOk {
				if needPointer(idVal.Name) {
					fld.Type = astPtrExpr(idVal.Name)
					logMsgf("Rewriting %s: making pointer var", idVal.Name)
					addOmitemptyTag(fld)
				}
			}

			// Array Type
			if atVal, atOk := fld.Type.(*ast.ArrayType); atOk {
				// If not a pointer
				if idVal, idOk := atVal.Elt.(*ast.Ident); idOk {
					if needPointer(idVal.Name) {
						atVal.Elt = astPtrExpr(idVal.Name)
						logMsgf("Rewriting %s: making pointer array", idVal.Name)
						addOmitemptyTag(fld)
					}
				}
			}

			// MapType
			if mtVal, mtOk := fld.Type.(*ast.MapType); mtOk {
				if idVal, idOk := mtVal.Value.(*ast.Ident); idOk {
					if needPointer(idVal.Name) {
						mtVal.Value = astPtrExpr(idVal.Name)
						logMsgf("Rewriting %s: making pointer map value", idVal.Name)
						addOmitemptyTag(fld)
					}
				}
			}

			// Is a simple pointer?
			if _, stOk := fld.Type.(*ast.StarExpr); stOk {
				addOmitemptyTag(fld)
			}

			// just append
			fl = append(fl, fld)
		}

		x.Fields.List = fl

		return x, true
	})

	format.Node(&buf, fset, f)

	if outFilename == "" {
		fmt.Println(buf.String())
		return
	}

	if err = ioutil.WriteFile(outFilename, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Could not write to outfile %s: %v", outFilename, err)
	}
}

func addOmitemptyTag(fld *ast.Field) {
	// Basic Literal. i.e simple string
	if blVal := fld.Tag; blVal != nil {
		re := regexp.MustCompile("`json:\"(.*?)\"")
		matches := re.FindAllStringSubmatch(blVal.Value, -1)
		if len(matches) > 0 && len(matches[0]) > 1 {
			orig := matches[0]
			if !strings.Contains(orig[1], "omitempty") {
				blVal.Value = strings.Replace(blVal.Value, orig[1], orig[1]+",omitempty", -1)
			}
		}
	}

}

func needPointer(typeName string) bool {
	return typesList.Contains(typeName)
}

func astPtrExpr(typeName string) *ast.StarExpr {
	return &ast.StarExpr{
		X: &ast.Ident{Name: typeName},
	}
}

func logMsgf(msg string, params ...interface{}) {
	fmt.Printf(msg, params...)
}
