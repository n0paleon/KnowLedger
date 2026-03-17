package utils

import (
	"strings"

	"gonum.org/v1/gonum/floats"
)

// CosineSimilarityText menghitung kemiripan teks dengan optimasi alokasi
func CosineSimilarityText(text1, text2 string) float64 {
	len1 := len(text1)
	len2 := len(text2)

	// --- Early Exit Logic ---
	// Jika perbedaan panjang karakter terlalu jauh, tidak perlu diproses.
	// Contoh: jika teks1 100 char dan teks2 500 char, ratio-nya 0.2.
	const minRatio = 0.5

	if len1 == 0 || len2 == 0 {
		return 0
	}

	ratio := float64(len1) / float64(len2)
	if len1 > len2 {
		ratio = float64(len2) / float64(len1)
	}

	if ratio < minRatio {
		return 0
	}

	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	allWords := make(map[string][2]float64, len(words1)+len(words2))

	for _, w := range words1 {
		w = strings.ToLower(w)
		entry := allWords[w]
		entry[0]++
		allWords[w] = entry
	}
	for _, w := range words2 {
		w = strings.ToLower(w)
		entry := allWords[w]
		entry[1]++
		allWords[w] = entry
	}

	vecA := make([]float64, 0, len(allWords))
	vecB := make([]float64, 0, len(allWords))

	for _, counts := range allWords {
		vecA = append(vecA, counts[0])
		vecB = append(vecB, counts[1])
	}

	dot := floats.Dot(vecA, vecB)
	normA := floats.Norm(vecA, 2)
	normB := floats.Norm(vecB, 2)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (normA * normB)
}
