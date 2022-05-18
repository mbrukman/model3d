package numerical

import (
	"math"
	"testing"
)

func TestMatrix4Mul(t *testing.T) {
	m1 := &Matrix4{-2.1130000, 1.4820000, 0.0370000, 0.3030000, 1.4960000, 0.3140000, 1.0620000, -0.8200000, -0.6650000, 0.5030000, -0.6730000, -0.7730000, 0.6070000, -0.4350000, 0.7850000, 2.0240000}
	m2 := &Matrix4{-0.4530000, 0.1840000, -1.0770000, 0.1830000, -0.3520000, 1.9300000, 0.4620000, 0.3640000, -1.0870000, -0.1670000, -0.5330000, -0.8320000, -1.3760000, 1.1500000, 2.0760000, 0.4800000}
	expected := &Matrix4{-0.0216220, 2.8137390, 3.5696920, 0.2674250, -0.8142900, -0.2390700, -3.7344900, -0.8891200, 1.9193880, 0.0718710, -0.2974480, 0.2502930, -3.7601700, 1.4686430, 2.9287100, 0.2711410}
	actual := m1.Mul(m2)

	for i, x := range expected {
		a := actual[i]
		if math.Abs(a-x) > 1e-7 {
			t.Errorf("entry %d: expected %f but got %f", i, x, a)
		}
	}
}

func TestMatrix4CharPoly(t *testing.T) {
	check := func(t *testing.T, actual, expected Polynomial) {
		if len(actual) != len(expected) {
			t.Fatalf("expected len %d but got %d", len(expected), len(actual))
		}
		for i, x := range expected {
			a := actual[i]
			if math.Abs(a-x) > 1e-3 {
				t.Errorf("term %d should be %f but got %f", i, x, a)
			}
		}
	}

	t.Run("NoThirdDegree", func(t *testing.T) {
		mat := &Matrix4{
			1.0, 3.0, 2.0, -7.0,
			5.0, -6.0, 3.0, -2.0,
			-4.0, 3.0, 2.0, 3.0,
			9.0, 1.0, 2.0, 3.0,
		}
		actual := mat.CharPoly()
		expected := Polynomial{-2420.0, 399.0, 18.0, 0.0, 1.0}
		check(t, actual, expected)
	})

	t.Run("Full", func(t *testing.T) {
		mat := &Matrix4{
			1.0, 3.0, 2.0, -7.0,
			5.0, -6.0, 3.0, -2.0,
			-4.0, 3.0, 2.0, 3.0,
			9.0, 1.0, 2.0, 4.0,
		}
		actual := mat.CharPoly()
		expected := Polynomial{-2525.0, 431.0, 15.0, -1.0, 1.0}
		check(t, actual, expected)
	})
}
