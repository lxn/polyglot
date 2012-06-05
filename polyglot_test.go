// Copyright 2012 The polyglot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polyglot

import (
	"testing"
)

func Test_NewDict(t *testing.T) {
	type test struct {
		locale       string
		expValidDict bool
		expErr       error
	}

	data := []test{
		{"de", true, nil},
		{"fr_FR", true, nil},
		{"EN", false, ErrInvalidLocale},
		{"EN_US", false, ErrInvalidLocale},
		{"_IT", false, ErrInvalidLocale},
		{"it_", false, ErrInvalidLocale},
		{"1de", false, ErrInvalidLocale},
		{"de_DE2", false, ErrInvalidLocale},
		{"_", false, ErrInvalidLocale},
		{"de-DE", false, ErrInvalidLocale},
	}

	for _, d := range data {
		dict, err := NewDict("testdata/basic", d.locale)

		if dict == nil && d.expValidDict {
			t.Errorf("Expected valid dict for locale '%s'", d.locale)
		}

		if err != d.expErr {
			if err == nil {
				t.Errorf("Expected '%v' error for locale '%s', but got nil", d.expErr, d.locale)
			} else if d.expErr == nil {
				t.Errorf("Expected no error for locale '%s', but got '%v'", d.locale, err)
			} else {
				t.Errorf("Expected '%v' error for locale '%s', but got '%v'", d.expErr, d.locale, err)
			}
		}
	}
}

func Test_Translation(t *testing.T) {
	type test struct {
		locale   string
		source   string
		context  []string
		expTrans string
	}

	data := []test{
		{"de", "Hello", nil, "Hallo"},
		{"de_DE", "Hello", nil, "Hallo"},
		{"es", "Hello", nil, "Hello"},
		{"de", "Hello", []string{"blah"}, "Hello"},
		{"de_DE", "Hello", []string{"blah"}, "Hello"},
		{"de", "Exit", []string{"noun"}, "Ausgang"},
		{"de_DE", "Exit", []string{"noun"}, "Ausgang"},
		{"de", "Exit", []string{"menu"}, "Beenden"},
		{"de_DE", "Exit", []string{"menu"}, "Beenden"},
		{"en", "Exit", []string{"menu"}, "Exit"},
		{"fr", "Exit", []string{"menu"}, "Exit"},
		{"it_IT", "Exit", []string{"menu"}, "Exit"},
	}

	for _, d := range data {
		dict, err := NewDict("testdata/basic", d.locale)

		if err != nil {
			t.Errorf("NewDict failed for locale '%s' with error %s", d.locale, err)
		}

		trans := dict.Translation(d.source, d.context...)

		if trans != d.expTrans {
			t.Errorf("Expected '%s' for locale '%s' and source '%s', but got '%s'", d.expTrans, d.locale, d.source, trans)
		}
	}
}
