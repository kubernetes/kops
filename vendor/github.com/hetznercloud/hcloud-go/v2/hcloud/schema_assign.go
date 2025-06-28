//go:build !goverter

package hcloud

/*
This file is needed so that c is assigned to a converterImpl{}.
If we did this in schema.go, goverter would fail because of a
compiler error (converterImpl might not be defined).
Because this file is not compiled by goverter, we can safely
assign c here.
*/

func init() {
	c = &converterImpl{}
}
