package e

import (
	"errors"
	"fmt"
	"testing"
)

func TestExternalError(t *testing.T) {

	type (
		Error1 struct{}
		Error2 int
	)

	var Ee = fmt.Errorf("This is an error")

	const (
		E1 ErrorStr[Error1] = "this is a test"
		E2 ErrorStr[Error1] = "this is a test: %w"
		E3 ErrorStr[Error2] = "this is a test"
	)

	type data struct {
		fn      func() error
		not     bool
		wantErr error
		wantStr string
	}

	test := func(test data) func(t *testing.T) {
		return func(t *testing.T) {
			have := test.fn()

			if test.wantErr != nil {
				if !errors.Is(have, test.wantErr) && !test.not {
					t.Errorf("\nhave: %#v\nwant: %#v\n", have, test.wantErr)
				}
			} else if test.wantStr == "" && have != nil {
				t.Errorf("have: %#v - is expected to be nil", have.Error())
			}

			if test.wantStr != "" {
				if have == nil {
					t.Errorf("have not expected to be nil")
				} else if have.Error() != test.wantStr {
					t.Errorf("\nhave: %q\nwant: %q\n", have.Error(), test.wantStr)
				}
			}

		}
	}

	t.Run("happy", test(data{
		fn: func() error {
			return E1
		},
		wantErr: E1,
	}))

	t.Run("nested error", test(data{
		fn: func() error {
			return E2.F(Ee)
		},
		wantErr: E2.F(Ee),
	}))

	t.Run("fwd mix nested error", test(data{
		fn: func() error {
			return E2.F(Ee)
		},
		wantErr: E2,
	}))

	t.Run("rev mix nested error", test(data{
		fn: func() error {
			return E2
		},
		not:     true,
		wantErr: E2.F(Ee),
	}))

	t.Run("fwd mix nested error diff types", test(data{
		fn: func() error {
			return E2.F(Ee)
		},
		not:     true,
		wantErr: E3,
	}))

	t.Run("error diff types", test(data{
		fn: func() error {
			return E1
		},
		not:     true,
		wantErr: E3,
	}))

	t.Run("kv", test(data{
		fn: func() error {
			return E2.KV("apple", "pie").F(Ee)
		},
		wantStr: `this is a test: This is an error {"apple":"pie"}`,
	}))

	t.Run("kv nested kv", test(data{
		fn: func() error {
			e2 := E2.KV("banana", "bread").F(Ee)
			return E2.KV("apple", "pie").F(e2)
		},
		wantStr: `this is a test: this is a test: This is an error {"apple":"pie","banana":"bread"}`,
	}))

	t.Run("nil", test(data{
		fn: func() error {
			return E2.KV("apple", "pie").F(nil)
		},
		wantStr: `this is a test: %!w(<nil>) {"apple":"pie"}`, // the nil is not nilable...
	}))

	t.Run("nil", test(data{
		fn: func() error {
			return E2.KV("apple", "pie").F(IsNilableErr(nil))
		},
		wantErr: nil, // it's nilable
	}))

	t.Run("nested error nilable", test(data{
		fn: func() error {
			return E2.F(IsNilableErr(Ee))
		},
		wantErr: E2.F(Ee),
	}))

	t.Run("nested error diff type", test(data{
		fn: func() error {
			e2 := E3.KV("banana", "bread")
			return E2.KV("apple", "pie").F(e2)
		},
		wantStr: `this is a test: this is a test {"apple":"pie","banana":"bread"}`,
	}))

	t.Run("nested nilable error diff type", test(data{
		fn: func() error {
			e2 := E3.KV("banana", "bread")
			return E2.KV("apple", "pie").F(IsNilableErr(e2))
		},
		wantStr: `this is a test: this is a test {"apple":"pie","banana":"bread"}`,
	}))
}
