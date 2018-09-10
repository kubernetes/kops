package govalidator

import "testing"

func BenchmarkContains(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Contains("a0b01c012deffghijklmnopqrstu0123456vwxyz", "0123456789")
	}
}

func BenchmarkMatches(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Matches("alfkjl12309fdjldfsa209jlksdfjLAKJjs9uJH234", "[\\w\\d]+")
	}
}
