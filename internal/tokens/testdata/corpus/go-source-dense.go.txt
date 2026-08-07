package vecx

type Num interface{ ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64 }

type Pair[K comparable, V any] struct{ K K; V V }

func Zip[K comparable, V any](ks []K, vs []V) []Pair[K, V] {
	n := len(ks)
	if m := len(vs); m < n {
		n = m
	}
	out := make([]Pair[K, V], 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Pair[K, V]{K: ks[i], V: vs[i]})
	}
	return out
}

func Fold[T, A any](xs []T, z A, f func(A, T) A) A {
	for _, x := range xs {
		z = f(z, x)
	}
	return z
}

func Dot[T Num](a, b []T) T {
	var s T
	n := min(len(a), len(b))
	i := 0
	for ; i+4 <= n; i += 4 {
		s += a[i]*b[i] + a[i+1]*b[i+1] + a[i+2]*b[i+2] + a[i+3]*b[i+3]
	}
	for ; i < n; i++ {
		s += a[i] * b[i]
	}
	return s
}

func Clamp[T Num](v, lo, hi T) T { return max(lo, min(hi, v)) }

func Chunk[T any](xs []T, n int) [][]T {
	if n <= 0 {
		return nil
	}
	out := make([][]T, 0, (len(xs)+n-1)/n)
	for i := 0; i < len(xs); i += n {
		j := i + n
		if j > len(xs) {
			j = len(xs)
		}
		out = append(out, xs[i:j:j])
	}
	return out
}

func Bits(x uint64) (n int) {
	for ; x != 0; x &= x - 1 {
		n++
	}
	return
}

func Rot(x uint32, k uint) uint32 { return x<<k | x>>(32-k) }

func Mix(h, v uint64) uint64 {
	h ^= v * 0x9e3779b97f4a7c15
	h = (h ^ (h >> 31)) * 0xbf58476d1ce4e5b9
	return h ^ (h >> 29)
}

func Idx(w, h, x, y int) int { return y*w + x }

func Norm(xs []float64) {
	var s float64
	for _, x := range xs {
		s += x * x
	}
	if s == 0 {
		return
	}
	inv := 1 / sqrt(s)
	for i := range xs {
		xs[i] *= inv
	}
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 8; i++ {
		z = (z + x/z) / 2
	}
	return z
}
