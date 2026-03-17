package utils

import (
	"math"
	"testing"
)

func TestCosineSimilarityText(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
	}{
		{
			name:     "Teks Identik",
			text1:    "golang itu keren",
			text2:    "golang itu keren",
			expected: 1.0,
		},
		{
			name:     "Teks Identik (Beda Case)",
			text1:    "GOLANG itu KEREN",
			text2:    "golang itu keren",
			expected: 1.0,
		},
		{
			name:     "Teks Berbeda Total",
			text1:    "apel merah manis",
			text2:    "mobil balap kencang",
			expected: 0.0,
		},
		{
			name:     "Teks Kosong",
			text1:    "",
			text2:    "ada isinya",
			expected: 0.0,
		},
		{
			name:  "Kemiripan Parsial",
			text1: "saya suka makan nasi",
			text2: "saya suka makan ayam",
			// Total kata unik: saya, suka, makan, nasi, ayam (5 dimensi)
			// VecA: [1, 1, 1, 1, 0], VecB: [1, 1, 1, 0, 1]
			// Dot: 3, NormA: sqrt(4), NormB: sqrt(4) -> 3 / (2 * 2) = 0.75
			expected: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarityText(tt.text1, tt.text2)

			if math.Abs(got-tt.expected) > 1e-4 {
				t.Errorf("%s: diharapkan %f, didapat %f", tt.name, tt.expected, got)
			}
		})
	}
}

func BenchmarkCosineSimilarityText(b *testing.B) {
	t1 := "Ini adalah konten teks yang cukup panjang untuk diuji performanya dalam sistem Go."
	t2 := "Ini adalah konten teks yang hampir sama untuk diuji performanya dalam sistem Go."

	for i := 0; i < b.N; i++ {
		CosineSimilarityText(t1, t2)
	}
}
