package dwnld

import (
	"fmt"
	"io"
	"time"
)

func print(writer io.Writer, rs resource, pad int) {
	switch rs.status {
	case startingCode:
		fmt.Fprintf(writer, "%s\n", printStarting(rs))
	case dwnldingCode:
		fmt.Fprintf(writer, "%s\n", printDwnld(rs, pad))
	case completeCode:
		fmt.Fprintf(writer, "%s\n", printComplete(rs, pad))
	default:
		fmt.Fprintf(writer, "%s\n", printError(rs))
	}
}

func printStarting(rs resource) string {
	tl := 33
	return fmt.Sprintf("[%*s] Waiting for download ", tl, truncate(rs.url, tl))
}

func printDwnld(rs resource, pad int) string {
	tl := 33

	header := fmt.Sprintf("[%*s] %-*s:", tl, truncate(rs.url, tl), pad, rs.name)

	dur := time.Since(rs.start)

	if rs.size <= 0 {

		return fmt.Sprintf("%s Downloading %s [start %s]", header, printByte(rs.written), formatDuration(dur))
	} else {
		remaining := time.Duration(0)
		if rs.written != 0 {
			remaining = time.Duration((float64(dur) / float64(rs.written)) * float64(rs.size-rs.written))
		}
		return fmt.Sprintf("%s Downloading %s / %s [etr %s]", header, printByte(rs.written), printByte(rs.size), formatDuration(remaining))
	}

}

func printComplete(rs resource, pad int) string {
	tl := 33
	header := fmt.Sprintf("[%*s] %-*s:", tl, truncate(rs.url, tl), pad, rs.name)

	// If the download is too fast it is possible that the start has not been set
	dur := time.Duration(0)
	if rs.start.Equal(time.Time{}) {
		dur = rs.end.Sub(rs.start)
	}

	return fmt.Sprintf("%s Download complete [in %s]", header, formatDuration(dur))
}

func printError(rs resource) string {
	if rs.name == "" {
		return fmt.Sprintf("%s: error - %s", rs.url, rs.err.Error())
	}
	return fmt.Sprintf("%s: error - %s", rs.name, rs.err.Error())
}

func truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length-3] + "..."
}

func printByte(b int64) string {
	return fmt.Sprintf("%*s", 8, byteNbToString(b))
}

func byteNbToString(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	mb := kb / 1024
	if mb < 1024 {
		return fmt.Sprintf("%.1fMB", mb)
	}
	gb := mb / 1024
	if gb < 1024 {
		return fmt.Sprintf("%.1fGB", gb)
	}
	tb := gb / 1024
	return fmt.Sprintf("%.1fTB", tb)
}

func formatDuration(t time.Duration) string {

	if t.Seconds() < 60 {
		return fmt.Sprintf("%.1fs", t.Seconds())
	}

	if t.Minutes() < 60 {
		m := int(t.Minutes())
		s := int((t.Minutes() - float64(m)) * 60)
		return fmt.Sprintf("%dm%ds", m, s)
	}

	h := int(t.Hours())
	m := int(t.Minutes() - float64((h * 60)))
	s := int(t.Seconds() - float64(int(t.Minutes()*60)))

	return fmt.Sprintf("%dh%dm%ds", h, m, s)
}
