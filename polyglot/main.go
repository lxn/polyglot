// Copyright 2012 The polyglot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"
)

var baseName *string = flag.String("name", "", "The base name to use for translation files.")
var directoryPath *string = flag.String("dir", "", "The directory path where to recursively search for Go files.")
var locales *string = flag.String("locales", "", `Comma-separated list of locales, for which to generate or update tr files. e.g.: "de_AT,de_DE,de,es,fr,it".`)

func logFatal(err error) {
	log.Fatalf(`An error occurred: %s
	
	Stack:
	%s`,
		err, debug.Stack())
}

type Location struct {
	File string
	Line string
}

type Message struct {
	Locations   []*Location
	Source      string
	Context     []string
	Translation string
}

type TRFile struct {
	Messages []*Message
}

func sourceKey(source string, context []string) string {
	if len(context) == 0 {
		return source
	}

	return fmt.Sprintf("__%s__%s__", source, strings.Join(context, "__"))
}

type visitor struct {
	fileSet           *token.FileSet
	sourceKey2Message map[string]*Message
}

func (v visitor) Visit(node ast.Node) (w ast.Visitor) {
	if callExpr, ok := node.(*ast.CallExpr); ok {
		if ident, ok := callExpr.Fun.(*ast.Ident); !ok || ident.Name != "tr" {
			return v
		}

		if len(callExpr.Args) > 0 {
			if basicLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
				pos := v.fileSet.Position(callExpr.Pos())

				source := string(basicLit.Value[1 : len(basicLit.Value)-1])
				var context []string
				for _, arg := range callExpr.Args[1:] {
					if basicLit, ok := arg.(*ast.BasicLit); ok {
						c := string(basicLit.Value[1 : len(basicLit.Value)-1])
						context = append(context, c)
					}
				}
				srcKey := sourceKey(source, context)
				message, ok := v.sourceKey2Message[srcKey]
				if !ok {
					message = &Message{Source: source, Context: context}
					v.sourceKey2Message[srcKey] = message
				}

				location := &Location{File: pos.Filename, Line: fmt.Sprintf("%d", pos.Line)}
				message.Locations = append(message.Locations, location)
			}
		}
	}

	return v
}

func (v visitor) scanDir(dirPath string) {
	dir, err := os.Open(dirPath)
	if err != nil {
		logFatal(err)
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		logFatal(err)
	}

	for _, name := range names {
		fullPath := path.Join(dirPath, name)

		fi, err := os.Stat(fullPath)
		if err != nil {
			logFatal(err)
		}

		if fi.IsDir() {
			v.scanDir(fullPath)
		} else if !fi.IsDir() && strings.HasSuffix(fullPath, ".go") {
			astFile, err := parser.ParseFile(v.fileSet, fullPath, nil, 0)
			if err != nil {
				logFatal(err)
			}

			ast.Walk(v, astFile)
		}
	}
}

func readOldSourceKey2MessageFromTRFile(filePath string) map[string]*Message {
	sk2m := make(map[string]*Message)

	if fi, _ := os.Stat(filePath); fi == nil {
		return sk2m
	}

	file, err := os.Open(filePath)
	if err != nil {
		logFatal(err)
	}
	defer file.Close()

	var trf TRFile
	if err := json.NewDecoder(file).Decode(&trf); err != nil {
		logFatal(err)
	}

	for _, msg := range trf.Messages {
		sk2m[sourceKey(msg.Source, msg.Context)] = msg
	}

	return sk2m
}

func writeTRFile(filePath string, sourceKey2Message, oldSourceKey2Message map[string]*Message, loc string) {
	file, err := os.Create(filePath)
	if err != nil {
		logFatal(err)
	}
	defer file.Close()

	var trf TRFile
	for _, msg := range sourceKey2Message {
		if oldMsg, ok := oldSourceKey2Message[sourceKey(msg.Source, msg.Context)]; ok {
			msg.Translation = oldMsg.Translation
		}

		trf.Messages = append(trf.Messages, msg)
	}

	if err := json.NewEncoder(file).Encode(trf); err != nil {
		logFatal(err)
	}
}

func main() {
	flag.Parse()

	if *baseName == "" || *directoryPath == "" || *locales == "" {
		flag.Usage()
		os.Exit(1)
	}

	v := visitor{
		fileSet:           token.NewFileSet(),
		sourceKey2Message: make(map[string]*Message),
	}

	v.scanDir(*directoryPath)

	locs := strings.Split(*locales, ",")
	for _, loc := range locs {
		loc = strings.TrimSpace(loc)

		filePath := fmt.Sprintf("%s-%s.tr", *baseName, loc)

		oldSourceKey2Message := readOldSourceKey2MessageFromTRFile(filePath)
		writeTRFile(filePath, v.sourceKey2Message, oldSourceKey2Message, loc)
	}
}
