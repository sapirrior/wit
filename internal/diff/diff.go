package diff

import (
	"fmt"
	"strings"
)

type Edit struct {
	Op   rune // '+', '-', ' '
	Line string
}

// Myers implements Myers' greedy diff algorithm.
func Myers(a, b []string) []Edit {
	N := len(a)
	M := len(b)
	MAX := N + M

	if N == 0 {
		edits := make([]Edit, M)
		for i, line := range b {
			edits[i] = Edit{Op: '+', Line: line}
		}
		return edits
	}
	if M == 0 {
		edits := make([]Edit, N)
		for i, line := range a {
			edits[i] = Edit{Op: '-', Line: line}
		}
		return edits
	}

	V := make([]int, 2*MAX+1)
	offset := MAX

	history := make([][]int, 0)
	found := false
	var d int
	for d = 0; d <= MAX; d++ {
		vCopy := make([]int, len(V))
		copy(vCopy, V)
		history = append(history, vCopy)

		for k := -d; k <= d; k += 2 {
			var x int
			kIdx := k + offset
			if k == -d || (k != d && V[kIdx-1] < V[kIdx+1]) {
				x = V[kIdx+1]
			} else {
				x = V[kIdx-1] + 1
			}
			y := x - k

			for x < N && y < M && a[x] == b[y] {
				x++
				y++
			}
			V[kIdx] = x

			if x >= N && y >= M {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	type Point struct{ X, Y int }
	path := []Point{}

	x := N
	y := M
	for d >= 0 {
		path = append(path, Point{X: x, Y: y})
		if d == 0 {
			break
		}
		v := history[d]
		k := x - y
		kIdx := k + offset

		var prevK int
		if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := v[prevK+offset]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			x--
			y--
			path = append(path, Point{X: x, Y: y})
		}

		x = prevX
		y = prevY
		d--
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	edits := []Edit{}
	currX, currY := 0, 0
	for _, p := range path {
		for currX < p.X && currY < p.Y && a[currX] == b[currY] {
			edits = append(edits, Edit{Op: ' ', Line: a[currX]})
			currX++
			currY++
		}
		for currX < p.X {
			edits = append(edits, Edit{Op: '-', Line: a[currX]})
			currX++
		}
		for currY < p.Y {
			edits = append(edits, Edit{Op: '+', Line: b[currY]})
			currY++
		}
	}

	return edits
}

// FormatUnified formats a list of Edits to a unified diff format.
func FormatUnified(fileName string, edits []Edit) string {
	var sb strings.Builder
	hasChanges := false
	for _, e := range edits {
		if e.Op != ' ' {
			hasChanges = true
			break
		}
	}
	if !hasChanges {
		return ""
	}

	sb.WriteString(fmt.Sprintf("--- %s\n", fileName))
	sb.WriteString(fmt.Sprintf("+++ %s\n", fileName))

	i := 0
	n := len(edits)
	contextSize := 3

	for i < n {
		for i < n && edits[i].Op == ' ' {
			i++
		}
		if i == n {
			break
		}

		hunkStart := i - contextSize
		if hunkStart < 0 {
			hunkStart = 0
		}

		hunkEnd := i
		consecutiveUnchanged := 0
		for hunkEnd < n {
			if edits[hunkEnd].Op != ' ' {
				consecutiveUnchanged = 0
			} else {
				consecutiveUnchanged++
				if consecutiveUnchanged > 2*contextSize {
					break
				}
			}
			hunkEnd++
		}

		hunkEndActual := hunkEnd
		if consecutiveUnchanged > 2*contextSize {
			hunkEndActual = hunkEnd - (consecutiveUnchanged - contextSize)
		}

		startA := 1
		startB := 1
		for j := 0; j < hunkStart; j++ {
			if edits[j].Op == ' ' {
				startA++
				startB++
			} else if edits[j].Op == '-' {
				startA++
			} else if edits[j].Op == '+' {
				startB++
			}
		}

		countA := 0
		countB := 0
		for j := hunkStart; j < hunkEndActual; j++ {
			if edits[j].Op == ' ' {
				countA++
				countB++
			} else if edits[j].Op == '-' {
				countA++
			} else if edits[j].Op == '+' {
				countB++
			}
		}

		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", startA, countA, startB, countB))
		for j := hunkStart; j < hunkEndActual; j++ {
			sb.WriteString(fmt.Sprintf("%c%s\n", edits[j].Op, edits[j].Line))
		}

		i = hunkEndActual
	}

	return sb.String()
}
