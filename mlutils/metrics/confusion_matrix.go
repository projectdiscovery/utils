package metrics

import "fmt"

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

func (cm *ConfusionMatrix) PrintConfusionMatrix() {
	fmt.Printf("%30s\n", "Confusion Matrix")
	fmt.Println()
	// Print header
	fmt.Printf("%-15s", "")
	for _, label := range cm.labels {
		fmt.Printf("%-15s", label)
	}
	fmt.Println()

	// Print rows
	for i, row := range cm.matrix {
		fmt.Printf("%-15s", cm.labels[i])
		for _, value := range row {
			fmt.Printf("%-15d", value)
		}
		fmt.Println()
	}
	fmt.Println()
}
