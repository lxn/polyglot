About polyglot
==============

polyglot is a String translation package and tool for Go.

Setup
=====

Make sure you have a working Go installation.
See [Getting Started](http://golang.org/doc/install.html)

Now run `go get github.com/lxn/polyglot` and 
`go get github.com/lxn/polyglot/polyglot`.

How does it work?
=================

1. It's pretty simple. Wrap translatable strings in your code in a call to a
   `func tr(source string, context ...string) string`, e.g. `tr("bla")`. You
   have to provide this function for every package you wish to use polyglot
   from.
2. After adding new translatable strings to your code, run the polyglot command,
   which scans your Go code for calls to a `tr` function. It will create or
   update JSON .tr files, adding new translatable strings that it finds.
3. Translate the strings.

Please see the hellopolyglot example for more details.