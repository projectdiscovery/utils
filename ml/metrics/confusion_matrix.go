package metrics

import (
	"fmt"
	"strings"
)

type ConfusionMatrix struct {
	matrix [][]int
	labels []string
}

func NewConfusionMatrix(actual, predicted []string, labels []string) *ConfusionMatrix {
	n := len(labels)
	matrix := make([][]int, n)
	for i := range matrix {
		matrix[i] = make([]int, n)
	}

	labelIndices := make(map[string]int)
	for i, label := range labels {
		labelIndices[label] = i
	}

	for i := range actual {
		matrix[labelIndices[actual[i]]][labelIndices[predicted[i]]]++
	}

	return &ConfusionMatrix{
		matrix: matrix,
		labels: labels,
	}
}

func (cm *ConfusionMatrix) PrintConfusionMatrix() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%30s\n", "Confusion Matrix")
	fmt.Fprintln(&s)
	// Print header
	fmt.Fprintf(&s, "%-15s", "")
	for _, label := range cm.labels {
		fmt.Fprintf(&s, "%-15s", label)
	}
	fmt.Fprintln(&s)

	// Print rows
	for i, row := range cm.matrix {
		fmt.Fprintf(&s, "%-15s", cm.labels[i])
		for _, value := range row {
			fmt.Fprintf(&s, "%-15d", value)
		}
		fmt.Fprintln(&s)
	}
	fmt.Fprintln(&s)

	return s.String()
}
