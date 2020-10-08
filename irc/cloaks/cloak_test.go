// Copyright (c) 2019 Shivaram Lingamneni
// released under the MIT license

package cloaks

import (
	"net"
	"reflect"
	"testing"
)

func assertEqual(supplied, expected interface{}, t *testing.T) {
	if !reflect.DeepEqual(supplied, expected) {
		t.Errorf("expected %v but got %v", expected, supplied)
	}
}

func easyParseIP(ipstr string) (result net.IP) {
	result = net.ParseIP(ipstr)
	if result == nil {
		panic(ipstr)
	}
	return
}

func cloakConfForTesting() CloakConfig {
	config := CloakConfig{
		Enabled:     true,
		Netname:     "oragono",
		secret:      "_BdVPWB5sray7McbFmeuJL996yaLgG4l9tEyficGXKg",
		CidrLenIPv4: 32,
		CidrLenIPv6: 64,
		NumBits:     80,
	}
	config.Initialize()
	return config
}

func TestCloakDeterminism(t *testing.T) {
	config := cloakConfForTesting()

	v4ip := easyParseIP("8.8.8.8").To4()
	assertEqual(config.ComputeCloak(v4ip), "d2z5guriqhzwazyr.oragono", t)
	// use of the 4-in-6 mapping should not affect the cloak
	v6mappedIP := v4ip.To16()
	assertEqual(config.ComputeCloak(v6mappedIP), "d2z5guriqhzwazyr.oragono", t)

	v6ip := easyParseIP("2001:0db8::1")
	assertEqual(config.ComputeCloak(v6ip), "w7ren6nxii6f3i3d.oragono", t)
	// same CIDR, so same cloak:
	v6ipsamecidr := easyParseIP("2001:0db8::2")
	assertEqual(config.ComputeCloak(v6ipsamecidr), "w7ren6nxii6f3i3d.oragono", t)
	v6ipdifferentcidr := easyParseIP("2001:0db9::1")
	// different CIDR, different cloak:
	assertEqual(config.ComputeCloak(v6ipdifferentcidr), "ccmptyrjwsxv4f4d.oragono", t)

	// cloak values must be sensitive to changes in the secret key
	config.SetSecret("HJcXK4lLawxBE4-9SIdPji_21YiL3N5r5f5-SPNrGVY")
	assertEqual(config.ComputeCloak(v4ip), "4khy3usk8mfu42pe.oragono", t)
	assertEqual(config.ComputeCloak(v6mappedIP), "4khy3usk8mfu42pe.oragono", t)
	assertEqual(config.ComputeCloak(v6ip), "mxpk3c83vdxkek9j.oragono", t)
	assertEqual(config.ComputeCloak(v6ipsamecidr), "mxpk3c83vdxkek9j.oragono", t)
}

func TestCloakShortv4Cidr(t *testing.T) {
	config := CloakConfig{
		Enabled:     true,
		Netname:     "oragono",
		secret:      "_BdVPWB5sray7McbFmeuJL996yaLgG4l9tEyficGXKg",
		CidrLenIPv4: 24,
		CidrLenIPv6: 64,
		NumBits:     60,
	}
	config.Initialize()

	v4ip := easyParseIP("8.8.8.8")
	assertEqual(config.ComputeCloak(v4ip), "3cay3zc72tnui.oragono", t)
	v4ipsamecidr := easyParseIP("8.8.8.9")
	assertEqual(config.ComputeCloak(v4ipsamecidr), "3cay3zc72tnui.oragono", t)
}

func TestCloakZeroBits(t *testing.T) {
	config := cloakConfForTesting()
	config.NumBits = 0
	config.Netname = "example.com"
	config.Initialize()

	v4ip := easyParseIP("8.8.8.8").To4()
	assertEqual(config.ComputeCloak(v4ip), "example.com", t)
}

func TestCloakDisabled(t *testing.T) {
	config := cloakConfForTesting()
	config.Enabled = false
	v4ip := easyParseIP("8.8.8.8").To4()
	assertEqual(config.ComputeCloak(v4ip), "", t)
}

func BenchmarkCloaks(b *testing.B) {
	config := cloakConfForTesting()
	v6ip := easyParseIP("2001:0db8::1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ComputeCloak(v6ip)
	}
}

func TestAccountCloak(t *testing.T) {
	config := cloakConfForTesting()

	// just assert that we get all distinct values
	assertEqual(config.ComputeAccountCloak("shivaram"), "8yu8kunudb45ztxm.oragono", t)
	assertEqual(config.ComputeAccountCloak("dolph🐬n"), "hhgeqsvzeagv3wjw.oragono", t)
	assertEqual(config.ComputeAccountCloak("SHIVARAM"), "bgx32x4r7qzih4uh.oragono", t)
	assertEqual(config.ComputeAccountCloak("ed"), "j5autmgxtdjdyzf4.oragono", t)
}

func TestAccountCloakCollisions(t *testing.T) {
	config := cloakConfForTesting()

	v4ip := easyParseIP("97.97.97.97")
	v4cloak := config.ComputeCloak(v4ip)
	// "aaaa" is the same bytestring as 97.97.97.97
	aaaacloak := config.ComputeAccountCloak("aaaa")
	if v4cloak == aaaacloak {
		t.Errorf("cloak collision between 97.97.97.97 and aaaa: %s", v4cloak)
	}
}
