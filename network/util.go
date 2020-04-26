package network

import "fmt"

func prettyBytes(bytes uint64) string {
	b := float64(bytes)
	if b < 1024 {
		return fmt.Sprintf("%.2f B", b)
	}

	b /= 1024
	if b < 1024 {
		return fmt.Sprintf("%.2f KiB", b)
	}

	b /= 1024
	if b < 1024 {
		return fmt.Sprintf("%.2f MiB", b)
	}

	b /= 1024
	return fmt.Sprintf("%.2f GiB", b)
}
