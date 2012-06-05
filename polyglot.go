// Copyright 2012 The polyglot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package polyglot provides a simple string translation mechanism.
package polyglot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var (
	// ErrInvalidLocale is returned if a specified locale is invalid.
	ErrInvalidLocale = errors.New("invalid locale")
)

type message struct {
	Source      string
	Context     []string
	Translation string
}

type trfile struct {
	Messages []*message
}

// Dict provides translated strings appropriate for a specific locale.
type Dict struct {
	dirPath                string
	locales                []string
	locale2SourceKey2Trans map[string]map[string]string
}

// NewDict returns a new Dict with the specified translations directory path and
// locale.
//
// The directory will be scanned recursively for JSON encoded .tr translation 
// files, as created by the polyglot tool, that have a name suffix matching one 
// of the locales in the locale chain.
// Example: Locale "en_US" has chain ["en_US", "en"], so files like
// foo-en_US.tr, foo-en.tr, bar-en.tr, baz-en.tr would be picked up.
func NewDict(translationsDirPath, locale string) (*Dict, error) {
	locales := localesChainForLocale(locale)
	if len(locales) == 0 {
		return nil, ErrInvalidLocale
	}

	d := &Dict{
		dirPath:                translationsDirPath,
		locales:                locales,
		locale2SourceKey2Trans: make(map[string]map[string]string),
	}

	if err := d.loadTranslations(translationsDirPath); err != nil {
		return nil, err
	}

	return d, nil
}

// DirPath returns the translations directory path of the Dict.
func (d *Dict) DirPath() string {
	return d.dirPath
}

// Locale returns the locale of the Dict.
func (d *Dict) Locale() string {
	return d.locales[0]
}

// Translation returns a translation of the source string to the locale of the 
// Dict or the source string, if no matching translation was found.
//
// Provided context arguments are used for disambiguation.
func (d *Dict) Translation(source string, context ...string) string {
	for _, locale := range d.locales {
		if sourceKey2Trans, ok := d.locale2SourceKey2Trans[locale]; ok {
			if trans, ok := sourceKey2Trans[sourceKey(source, context)]; ok {
				return trans
			}
		}
	}

	return source
}

func (d *Dict) loadTranslation(reader io.Reader, locale string) error {
	var trf trfile

	if err := json.NewDecoder(reader).Decode(&trf); err != nil {
		return err
	}

	sourceKey2Trans, ok := d.locale2SourceKey2Trans[locale]
	if !ok {
		sourceKey2Trans = make(map[string]string)

		d.locale2SourceKey2Trans[locale] = sourceKey2Trans
	}

	for _, m := range trf.Messages {
		if m.Translation != "" {
			sourceKey2Trans[sourceKey(m.Source, m.Context)] = m.Translation
		}
	}

	return nil
}

func (d *Dict) loadTranslations(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		fullPath := path.Join(dirPath, name)

		fi, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if err := d.loadTranslations(fullPath); err != nil {
				return err
			}
		} else if locale := d.matchingLocaleFromFileName(name); locale != "" {
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if err := d.loadTranslation(file, locale); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Dict) matchingLocaleFromFileName(name string) string {
	for _, locale := range d.locales {
		if strings.HasSuffix(name, fmt.Sprintf("-%s.tr", locale)) {
			return locale
		}
	}

	return ""
}

func sourceKey(source string, context []string) string {
	if len(context) == 0 {
		return source
	}

	return fmt.Sprintf("__%s__%s__", source, strings.Join(context, "__"))
}

func localesChainForLocale(locale string) []string {
	parts := strings.Split(locale, "_")
	if len(parts) > 2 {
		return nil
	}

	if len(parts[0]) != 2 {
		return nil
	}

	for _, r := range parts[0] {
		if r < rune('a') || r > rune('z') {
			return nil
		}
	}

	if len(parts) == 1 {
		return []string{parts[0]}
	}

	if len(parts[1]) < 2 || len(parts[1]) > 3 {
		return nil
	}

	for _, r := range parts[1] {
		if r < rune('A') || r > rune('Z') {
			return nil
		}
	}

	return []string{locale, parts[0]}
}
