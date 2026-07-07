package utils

func SplitTasks[T any](tasks []T, n int) [][]T {
	var chunks [][]T
	length := len(tasks)
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		start := i * length / n
		end := (i + 1) * length / n
		chunks = append(chunks, tasks[start:end])
	}
	return chunks
}
