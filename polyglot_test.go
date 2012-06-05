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
		dict, err := NewDict("testdata/hello", d.locale)

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
