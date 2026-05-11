package processing

import (
	"regexp"
	"strings"
)

var (
	uuidRegex      = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	hexUUIDRegex   = regexp.MustCompile(`[0-9a-fA-F]{32}`)
	hexAddrRegex   = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	numericIDRegex = regexp.MustCompile(`\b\d{5,}\b`)
	timestampRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[^\s]*`)
	emailRegex     = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	urlRegex       = regexp.MustCompile(`https?://[^\s"']+`)
	chunkHashRegex = regexp.MustCompile(`[.\-][0-9a-f]{8,}\.(js|css|mjs)`)
)

type NormalizedData struct {
	Message        string
	ExceptionValue string
	Frames         []NormalizedFrame
}

type NormalizedFrame struct {
	Filename string
	Function string
	Lineno   int
}

func Normalize(event *ParsedEvent) NormalizedData {
	nd := NormalizedData{}

	nd.Message = normalizeString(event.Message)

	if event.Exception != nil && len(event.Exception.Values) > 0 {
		last := event.Exception.Values[len(event.Exception.Values)-1]
		nd.ExceptionValue = normalizeString(last.Value)

		if last.Stacktrace != nil {
			frames := getInAppFrames(last.Stacktrace.Frames)
			for _, f := range frames {
				nd.Frames = append(nd.Frames, NormalizedFrame{
					Filename: normalizeFilename(f.Filename),
					Function: f.Function,
					Lineno:   f.Lineno,
				})
			}
		}
	}

	return nd
}

func normalizeString(s string) string {
	s = uuidRegex.ReplaceAllString(s, "<uuid>")
	s = hexUUIDRegex.ReplaceAllString(s, "<uuid>")
	s = timestampRegex.ReplaceAllString(s, "<timestamp>")
	s = hexAddrRegex.ReplaceAllString(s, "<addr>")
	s = emailRegex.ReplaceAllString(s, "<email>")
	s = urlRegex.ReplaceAllString(s, "<url>")
	s = numericIDRegex.ReplaceAllString(s, "<id>")
	s = strings.ToLower(strings.TrimSpace(s))
	return s
}

func normalizeFilename(filename string) string {
	filename = chunkHashRegex.ReplaceAllString(filename, ".<ext>")
	parts := strings.Split(filename, "?")
	filename = parts[0]
	parts = strings.Split(filename, "#")
	filename = parts[0]
	return filename
}

func getInAppFrames(frames []Frame) []Frame {
	var inApp []Frame
	for _, f := range frames {
		if f.InApp != nil && *f.InApp {
			inApp = append(inApp, f)
		}
	}
	if len(inApp) == 0 {
		return frames
	}
	return inApp
}
