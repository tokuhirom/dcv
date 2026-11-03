package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"

	"github.com/tokuhirom/dcv/internal/docker"
)

var (
	jsonKeyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	jsonStringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	jsonNumberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	jsonLiteralStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
)

// logSearchMatch identifies a single search hit within a log line.
type logSearchMatch struct {
	lineIndex int
	start     int // byte offset in line
	end       int // byte offset in line (exclusive)
}

// wrappedSegment is one display row of a log line.
type wrappedSegment struct {
	text      string
	startByte int
	endByte   int
}

type jsonColorRange struct {
	start int
	end   int
	style lipgloss.Style
}

func (m *LogViewModel) lineVisualCount(line string, effectiveWidth int) int {
	if !m.WrapText() {
		return max(strings.Count(line, "\n")+1, 1)
	}
	return len(wrapLineSegments(line, effectiveWidth))
}

// wrapLineSegments splits line into display-width rows for wrapped rendering.
func wrapLineSegments(line string, width int) []wrappedSegment {
	if width <= 0 || line == "" {
		if line == "" {
			return []wrappedSegment{{text: "", startByte: 0, endByte: 0}}
		}
		return []wrappedSegment{{text: line, startByte: 0, endByte: len(line)}}
	}

	var segments []wrappedSegment
	startByte := 0
	col := 0
	var b strings.Builder

	flush := func(endByte int) {
		if endByte <= startByte && len(segments) > 0 {
			return
		}
		segments = append(segments, wrappedSegment{
			text:      b.String(),
			startByte: startByte,
			endByte:   endByte,
		})
		b.Reset()
		startByte = endByte
		col = 0
	}

	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		if r == '\n' {
			segments = append(segments, wrappedSegment{
				text:      b.String(),
				startByte: startByte,
				endByte:   i,
			})
			b.Reset()
			startByte = i + size
			col = 0
			i += size
			continue
		}
		rw := runewidth.RuneWidth(r)
		if col > 0 && col+rw > width {
			flush(i)
		}
		b.WriteRune(r)
		col += rw
		i += size
	}

	flush(len(line))
	if len(segments) == 0 {
		return []wrappedSegment{{text: "", startByte: 0, endByte: 0}}
	}
	return segments
}

// displayWidthToByteIndex returns the byte index at the given display column.
func displayWidthToByteIndex(s string, targetCol int) int {
	if targetCol <= 0 {
		return 0
	}
	col := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		rw := runewidth.RuneWidth(r)
		if col >= targetCol {
			return i
		}
		col += rw
		i += size
	}
	return len(s)
}

func (m *LogViewModel) logLineSegments(line string, effectiveWidth int) []wrappedSegment {
	if strings.Contains(line, "\n") && !m.WrapText() {
		var segments []wrappedSegment
		offset := 0
		for _, part := range strings.SplitAfter(line, "\n") {
			text := strings.TrimSuffix(part, "\n")
			start := displayWidthToByteIndex(text, m.logScrollX)
			end := start
			col := 0
			for i := start; i < len(text) && col < effectiveWidth; {
				r, size := utf8.DecodeRuneInString(text[i:])
				rw := runewidth.RuneWidth(r)
				if col+rw > effectiveWidth {
					break
				}
				col += rw
				end = i + size
				i += size
			}
			segments = append(segments, wrappedSegment{
				text:      text[start:end],
				startByte: offset + start,
				endByte:   offset + end,
			})
			offset += len(part)
		}
		return segments
	}
	if strings.Contains(line, "\n") {
		return wrapLineSegments(line, effectiveWidth)
	}
	if m.WrapText() {
		return wrapLineSegments(line, effectiveWidth)
	}

	startByte := displayWidthToByteIndex(line, m.logScrollX)
	endByte := startByte
	col := 0
	for i := startByte; i < len(line) && col < effectiveWidth; {
		r, size := utf8.DecodeRuneInString(line[i:])
		rw := runewidth.RuneWidth(r)
		if col+rw > effectiveWidth {
			break
		}
		col += rw
		endByte = i + size
		i += size
	}

	return []wrappedSegment{{
		text:      line[startByte:endByte],
		startByte: startByte,
		endByte:   endByte,
	}}
}

func (m *LogViewModel) currentSearchMatch() (logSearchMatch, bool) {
	if len(m.logSearchMatches) == 0 || m.currentSearchIdx >= len(m.logSearchMatches) {
		return logSearchMatch{}, false
	}
	return m.logSearchMatches[m.currentSearchIdx], true
}

func (m *LogViewModel) isCurrentMatchOnSegment(lineIndex int, seg wrappedSegment) bool {
	current, ok := m.currentSearchMatch()
	if !ok || current.lineIndex != lineIndex {
		return false
	}
	return current.start >= seg.startByte && current.start < seg.endByte
}

func (m *LogViewModel) calculateMaxScrollX(model *Model, logsToDisplay []string) int {
	if m.WrapText() || len(logsToDisplay) == 0 {
		return 0
	}

	effectiveWidth := model.width - 2
	if effectiveWidth <= 0 {
		return 0
	}

	maxScrollX := 0
	for _, line := range logsToDisplay {
		for _, part := range strings.Split(line, "\n") {
			lineWidth := runewidth.StringWidth(part)
			if lineWidth > effectiveWidth {
				overflow := lineWidth - effectiveWidth
				if overflow > maxScrollX {
					maxScrollX = overflow
				}
			}
		}
	}

	return maxScrollX
}

type LogViewModel struct {
	SearchViewModel
	FilterViewModel

	logs       []string
	logScrollY int
	logScrollX int
	// logScrollVisual is the number of wrapped display rows skipped from the top.
	logScrollVisual int

	logSearchMatches []logSearchMatch

	container  *docker.Container
	prettyJSON bool

	LogReaderManager
}

func (m *LogViewModel) SwitchToLogView(model *Model, container *docker.Container) {
	model.SwitchView(LogView)

	m.container = container
	m.logs = []string{}
	m.logScrollY = 0
	m.logScrollX = 0
	m.logScrollVisual = 0
}

func (m *LogViewModel) StreamContainerLogs(model *Model, container *docker.Container) tea.Cmd {
	m.SwitchToLogView(model, container)
	args := container.OperationArgs("logs", "--tail", "1000", "--timestamps", "--follow")
	cmd := docker.Execute(args...)
	return m.streamLogsReal(cmd)
}

func (m *LogViewModel) HandleBack(model *Model) tea.Cmd {
	m.stopLogReader()
	model.SwitchToPreviousView()
	return nil
}

func (m *LogViewModel) displayLogs() []string {
	logs := m.logs
	if m.filterMode && m.filterText != "" {
		logs = m.filteredLogs
	} else if m.filterActive {
		logs = m.filteredLogs
	}

	if !m.prettyJSON {
		return logs
	}

	displayLogs := make([]string, len(logs))
	for i, line := range logs {
		displayLogs[i] = formatJSONLogLine(line)
	}
	return displayLogs
}

func splitJSONLogPrefix(line string) (string, string) {
	if token, rest, ok := strings.Cut(line, " "); ok {
		if _, err := time.Parse(time.RFC3339Nano, token); err == nil {
			return token + " ", rest
		}
	}
	return "", line
}

func formatJSONLogLine(line string) string {
	prefix, payload := splitJSONLogPrefix(line)
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(payload), "", "  "); err != nil {
		return line
	}
	return prefix + pretty.String()
}

func jsonLogColorRanges(line string) []jsonColorRange {
	prefix, payload := splitJSONLogPrefix(line)
	if !json.Valid([]byte(payload)) {
		return nil
	}

	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	var ranges []jsonColorRange
	for {
		start := int(decoder.InputOffset())
		token, err := decoder.Token()
		if err != nil {
			if err != io.EOF {
				slog.Error("Failed to color JSON log record", "error", err)
				return nil
			}
			return ranges
		}
		end := int(decoder.InputOffset())
		for start < end && strings.IndexByte(" \t\r\n,:", payload[start]) >= 0 {
			start++
		}

		var style lipgloss.Style
		switch token.(type) {
		case string:
			style = jsonStringStyle
			next := end
			for next < len(payload) && strings.IndexByte(" \t\r\n", payload[next]) >= 0 {
				next++
			}
			if next < len(payload) && payload[next] == ':' {
				style = jsonKeyStyle
			}
		case json.Number:
			style = jsonNumberStyle
		case bool, nil:
			style = jsonLiteralStyle
		default:
			continue
		}
		ranges = append(ranges, jsonColorRange{
			start: len(prefix) + start,
			end:   len(prefix) + end,
			style: style,
		})
	}
}

func renderJSONRange(line string, start, end int, ranges []jsonColorRange) string {
	if len(ranges) == 0 {
		return line[start:end]
	}
	var result strings.Builder
	pos := start
	for _, colorRange := range ranges {
		if colorRange.end <= pos {
			continue
		}
		if colorRange.start >= end {
			break
		}
		coloredStart := max(pos, colorRange.start)
		coloredEnd := min(end, colorRange.end)
		result.WriteString(line[pos:coloredStart])
		result.WriteString(colorRange.style.Render(line[coloredStart:coloredEnd]))
		pos = coloredEnd
	}
	result.WriteString(line[pos:end])
	return result.String()
}

func (m *LogViewModel) totalVisualLines(logs []string, effectiveWidth int) int {
	if !m.WrapText() {
		return len(logs)
	}
	total := 0
	for _, line := range logs {
		total += len(wrapLineSegments(line, effectiveWidth))
	}
	return total
}

func (m *LogViewModel) visualLineOffsetForMatch(logs []string, match logSearchMatch, effectiveWidth int) int {
	offset := 0
	for i := 0; i < match.lineIndex && i < len(logs); i++ {
		offset += len(wrapLineSegments(logs[i], effectiveWidth))
	}
	if match.lineIndex >= len(logs) {
		return offset
	}
	for _, seg := range wrapLineSegments(logs[match.lineIndex], effectiveWidth) {
		if match.start >= seg.startByte && match.start < seg.endByte {
			break
		}
		offset++
	}
	return offset
}

func (m *LogViewModel) calculateMaxVisualScroll(model *Model) int {
	logs := m.displayLogs()
	if len(logs) == 0 {
		return 0
	}

	effectiveWidth := model.width - 2
	if effectiveWidth <= 0 {
		effectiveWidth = 1
	}

	visibleHeight := model.PageSize() - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	total := m.totalVisualLines(logs, effectiveWidth)
	maxScroll := total - visibleHeight
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func (m *LogViewModel) clampVisualScroll(model *Model) {
	maxScroll := m.calculateMaxVisualScroll(model)
	if m.logScrollVisual > maxScroll {
		m.logScrollVisual = maxScroll
	}
	if m.logScrollVisual < 0 {
		m.logScrollVisual = 0
	}
}

func (m *LogViewModel) scrollToTop() {
	m.logScrollY = 0
	m.logScrollVisual = 0
}

func (m *LogViewModel) scrollToBottom(model *Model) {
	if m.WrapText() {
		m.logScrollVisual = m.calculateMaxVisualScroll(model)
		return
	}
	maxScroll := m.calculateMaxScroll(model)
	if maxScroll > 0 {
		m.logScrollY = maxScroll
	} else {
		m.logScrollY = 0
	}
}

func (m *LogViewModel) scrollBy(model *Model, delta int) {
	if m.WrapText() {
		m.logScrollVisual += delta
		if model != nil {
			m.clampVisualScroll(model)
		} else if m.logScrollVisual < 0 {
			m.logScrollVisual = 0
		}
		return
	}

	m.logScrollY += delta
	if m.logScrollY < 0 {
		m.logScrollY = 0
	}
	if model == nil {
		return
	}
	maxScroll := m.calculateMaxScroll(model)
	if m.logScrollY > maxScroll && maxScroll > 0 {
		m.logScrollY = maxScroll
	} else if maxScroll <= 0 {
		m.logScrollY = 0
	}
}

func (m *LogViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	if model.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	logsToDisplay := m.displayLogs()
	visibleHeight := availableHeight - 2
	effectiveWidth := model.width - 2
	if effectiveWidth <= 0 {
		effectiveWidth = 1
	}

	if len(logsToDisplay) == 0 {
		if (m.filterMode || m.filterActive) && m.filterText != "" {
			s.WriteString("No logs match the filter.\n")
		} else {
			s.WriteString("No logs available.\n")
		}
		return s.String()
	}

	if m.WrapText() {
		return m.renderWrapped(model, logsToDisplay, visibleHeight, effectiveWidth)
	}

	return m.renderNoWrap(model, logsToDisplay, visibleHeight, effectiveWidth)
}

func (m *LogViewModel) renderWrapped(model *Model, logsToDisplay []string, visibleHeight, effectiveWidth int) string {
	var s strings.Builder

	m.clampVisualScroll(model)
	skip := m.logScrollVisual
	rendered := 0
	startIdx := -1
	endIdx := 0

	for i, line := range logsToDisplay {
		var colorRanges []jsonColorRange
		colorsReady := false
		for _, seg := range wrapLineSegments(line, effectiveWidth) {
			if skip > 0 {
				skip--
				continue
			}
			if rendered >= visibleHeight {
				goto done
			}

			if startIdx < 0 {
				startIdx = i
			}
			endIdx = i + 1

			if !colorsReady {
				colorRanges = jsonLogColorRanges(line)
				colorsReady = true
			}
			displayLine := m.renderSegment(line, i, seg, colorRanges)

			if m.isCurrentMatchOnSegment(i, seg) {
				s.WriteString("> ")
			} else {
				s.WriteString("  ")
			}
			s.WriteString(displayLine + ResetAll + "\n")
			rendered++
		}
	}

done:
	if startIdx < 0 {
		startIdx = 0
	}

	m.appendScrollIndicator(&s, model, logsToDisplay, startIdx, endIdx, visibleHeight, effectiveWidth)
	return s.String()
}

func (m *LogViewModel) renderNoWrap(model *Model, logsToDisplay []string, visibleHeight, effectiveWidth int) string {
	var s strings.Builder

	startIdx := m.logScrollY
	endIdx := startIdx

	rendered := 0
	for i := startIdx; i < len(logsToDisplay) && rendered < visibleHeight; i++ {
		line := logsToDisplay[i]
		colorRanges := jsonLogColorRanges(line)
		for _, seg := range m.logLineSegments(line, effectiveWidth) {
			if rendered >= visibleHeight {
				break
			}
			displayLine := m.renderSegment(line, i, seg, colorRanges)

			if m.isCurrentMatchOnSegment(i, seg) {
				s.WriteString("> ")
			} else {
				s.WriteString("  ")
			}
			s.WriteString(displayLine + ResetAll + "\n")
			rendered++
			endIdx = i + 1
		}
	}

	m.appendScrollIndicator(&s, model, logsToDisplay, startIdx, endIdx, visibleHeight, effectiveWidth)
	return s.String()
}

func (m *LogViewModel) appendScrollIndicator(
	s *strings.Builder,
	model *Model,
	logsToDisplay []string,
	startIdx, endIdx, visibleHeight, effectiveWidth int,
) {
	totalVisual := m.totalVisualLines(logsToDisplay, effectiveWidth)
	needsIndicator := totalVisual > visibleHeight
	if !m.WrapText() {
		needsIndicator = needsIndicator || m.calculateMaxScrollX(model, logsToDisplay) > 0
	}
	if !needsIndicator {
		return
	}

	scrollInfo := fmt.Sprintf(" [%d-%d/%d] ", startIdx+1, endIdx, len(logsToDisplay))
	if (m.filterMode || m.filterActive) && m.filterText != "" {
		scrollInfo += fmt.Sprintf(" (filtered from %d)", len(m.logs))
	}
	if !m.WrapText() {
		maxScrollX := m.calculateMaxScrollX(model, logsToDisplay)
		if maxScrollX > 0 {
			scrollInfo += fmt.Sprintf(" col %d-%d/%d", m.logScrollX+1, m.logScrollX+effectiveWidth, maxScrollX+effectiveWidth)
		}
	}
	s.WriteString("\n" + helpStyle.Render(scrollInfo))
}

func (m *LogViewModel) findLogSearchMatches(logs []string) []logSearchMatch {
	if m.searchText == "" {
		return nil
	}

	var matches []logSearchMatch

	if m.searchRegex {
		pattern := m.searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil
		}
		for lineIndex, line := range logs {
			for _, match := range re.FindAllStringIndex(line, -1) {
				matches = append(matches, logSearchMatch{
					lineIndex: lineIndex,
					start:     match[0],
					end:       match[1],
				})
			}
		}
		return matches
	}

	searchStr := m.searchText
	if m.searchIgnoreCase {
		searchStr = strings.ToLower(searchStr)
	}
	searchLen := len(m.searchText)

	for lineIndex, line := range logs {
		lineToSearch := line
		if m.searchIgnoreCase {
			lineToSearch = strings.ToLower(line)
		}

		start := 0
		for {
			idx := strings.Index(lineToSearch[start:], searchStr)
			if idx == -1 {
				break
			}
			matchStart := start + idx
			matches = append(matches, logSearchMatch{
				lineIndex: lineIndex,
				start:     matchStart,
				end:       matchStart + searchLen,
			})
			start = matchStart + searchLen
		}
	}

	return matches
}

// PerformLogSearch finds every search occurrence and scrolls to the first hit.
func (m *LogViewModel) PerformLogSearch(model *Model) {
	m.logSearchMatches = m.findLogSearchMatches(m.displayLogs())
	m.searchResults = nil
	m.currentSearchIdx = 0

	if len(m.logSearchMatches) > 0 {
		m.scrollToSearchMatch(model, m.logSearchMatches[0])
	}
}

func (m *LogViewModel) scrollToSearchMatch(model *Model, match logSearchMatch) {
	logs := m.displayLogs()
	effectiveWidth := model.width - 2
	if effectiveWidth <= 0 {
		effectiveWidth = 1
	}

	if m.WrapText() {
		offset := m.visualLineOffsetForMatch(logs, match, effectiveWidth)
		visibleHeight := model.PageSize() - 2
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		centered := offset - visibleHeight/2
		if centered < 0 {
			centered = 0
		}
		m.logScrollVisual = centered
		m.clampVisualScroll(model)
		return
	}

	m.logScrollY = match.lineIndex - model.Height/2 + 3
	if m.logScrollY < 0 {
		m.logScrollY = 0
	}
	maxScroll := m.calculateMaxScroll(model)
	if m.logScrollY > maxScroll {
		m.logScrollY = maxScroll
	}
}

// renderSegment returns the display text for a single segment, applying
// filter or search highlighting where appropriate. Operating on raw bytes
// here avoids slicing through ANSI escapes added by highlighting helpers.
func (m *LogViewModel) renderSegment(fullLine string, lineIndex int, seg wrappedSegment, colorRanges []jsonColorRange) string {
	if (m.filterMode || m.filterActive) && m.filterText != "" {
		return m.highlightFilterSegment(fullLine, seg, colorRanges)
	}
	if m.searchText != "" && !m.searchMode {
		return m.highlightSegment(fullLine, lineIndex, seg, colorRanges)
	}
	return renderJSONRange(fullLine, seg.startByte, seg.endByte, colorRanges)
}

func (m *LogViewModel) highlightSegment(fullLine string, lineIndex int, seg wrappedSegment, colorRanges []jsonColorRange) string {
	text := fullLine[seg.startByte:seg.endByte]
	if m.searchText == "" {
		return renderJSONRange(fullLine, seg.startByte, seg.endByte, colorRanges)
	}

	current, hasCurrent := m.currentSearchMatch()
	matchRanges := m.matchRangesOnLine(fullLine, lineIndex)
	if len(matchRanges) == 0 {
		return renderJSONRange(fullLine, seg.startByte, seg.endByte, colorRanges)
	}

	var result strings.Builder
	lastEnd := 0
	for _, match := range matchRanges {
		if match.end <= seg.startByte || match.start >= seg.endByte {
			continue
		}

		relStart := max(match.start-seg.startByte, 0)
		relEnd := min(match.end-seg.startByte, len(text))
		if relStart > lastEnd {
			result.WriteString(renderJSONRange(fullLine, seg.startByte+lastEnd, seg.startByte+relStart, colorRanges))
		}

		style := searchMatchStyle
		if hasCurrent && current.lineIndex == lineIndex && current.start == match.start {
			style = searchCurrentMatchStyle
		}
		result.WriteString(style.Render(text[relStart:relEnd]))
		lastEnd = relEnd
	}
	if lastEnd < len(text) {
		result.WriteString(renderJSONRange(fullLine, seg.startByte+lastEnd, seg.endByte, colorRanges))
	}
	return result.String()
}

func (m *LogViewModel) matchRangesOnLine(line string, lineIndex int) []logSearchMatch {
	var ranges []logSearchMatch
	for _, match := range m.logSearchMatches {
		if match.lineIndex == lineIndex {
			ranges = append(ranges, match)
		}
	}
	return ranges
}

// highlightFilterSegment highlights filter matches within a single display
// segment. Matches are computed against fullLine (raw, ANSI-free) and rendered
// only for the portion that overlaps seg, so callers can slice safely without
// risk of cutting an ANSI escape in half.
func (m *LogViewModel) highlightFilterSegment(fullLine string, seg wrappedSegment, colorRanges []jsonColorRange) string {
	text := fullLine[seg.startByte:seg.endByte]
	if m.filterText == "" {
		return renderJSONRange(fullLine, seg.startByte, seg.endByte, colorRanges)
	}

	searchStr := strings.ToLower(m.filterText)
	lineToSearch := strings.ToLower(fullLine)
	searchLen := len(searchStr)

	var result strings.Builder
	lastEnd := 0
	scan := 0
	for scan < len(fullLine) {
		idx := strings.Index(lineToSearch[scan:], searchStr)
		if idx == -1 {
			break
		}
		matchStart := scan + idx
		matchEnd := matchStart + searchLen
		scan = matchEnd

		if matchEnd <= seg.startByte || matchStart >= seg.endByte {
			continue
		}

		relStart := max(matchStart-seg.startByte, 0)
		relEnd := min(matchEnd-seg.startByte, len(text))
		if relStart > lastEnd {
			result.WriteString(renderJSONRange(fullLine, seg.startByte+lastEnd, seg.startByte+relStart, colorRanges))
		}
		result.WriteString(searchMatchStyle.Render(text[relStart:relEnd]))
		lastEnd = relEnd
	}
	if lastEnd < len(text) {
		result.WriteString(renderJSONRange(fullLine, seg.startByte+lastEnd, seg.endByte, colorRanges))
	}
	return result.String()
}

// calculateMaxScroll calculates the maximum scroll position accounting for wrapped lines
func (m *LogViewModel) calculateMaxScroll(model *Model) int {
	logsToDisplay := m.displayLogs()

	if len(logsToDisplay) == 0 {
		return 0
	}

	// Must match the visibleHeight used in render(): availableHeight - 2
	// PageSize() returns Height - LogViewChromeOffset, which corresponds to
	// the availableHeight passed to render(). Render then subtracts 2 more.
	visibleHeight := model.PageSize() - 2

	// Work backwards from the last line accumulating visual lines.
	// When adding a line would exceed visibleHeight, the max scroll
	// is the next line (i+1) — all lines from i+1 to end fit on screen.
	// Each log line is prefixed with "  " or "> " (2 chars)
	effectiveWidth := model.width - 2
	visualLinesFromEnd := 0
	maxScroll := 0

	for i := len(logsToDisplay) - 1; i >= 0; i-- {
		lineVisualLines := m.lineVisualCount(logsToDisplay[i], effectiveWidth)
		visualLinesFromEnd += lineVisualLines
		if visualLinesFromEnd > visibleHeight {
			maxScroll = i + 1
			// Cap at last valid index — if only the last line exceeds
			// visibleHeight, we still want to be able to scroll to it.
			if maxScroll >= len(logsToDisplay) {
				maxScroll = len(logsToDisplay) - 1
			}
			break
		}
	}

	if maxScroll == 0 {
		// All content fits on screen, no scrolling needed
		return 0
	}

	// Ensure every line from maxScroll to end is actually visible in
	// the render by walking forward. The render always includes the
	// first line, then subsequent lines only if they fit. If some line
	// would be skipped, increase maxScroll so that line becomes the
	// first line shown.
	visualLinesUsed := 0
	for i := maxScroll; i < len(logsToDisplay); i++ {
		lineVisualLines := m.lineVisualCount(logsToDisplay[i], effectiveWidth)
		if i == maxScroll || visualLinesUsed+lineVisualLines <= visibleHeight {
			visualLinesUsed += lineVisualLines
		} else {
			// This line won't be shown from maxScroll, so allow scrolling further
			return len(logsToDisplay) - 1
		}
	}

	return maxScroll
}

func (m *LogViewModel) HandleUp() tea.Cmd {
	if m.WrapText() {
		m.scrollBy(nil, -1)
	} else if m.logScrollY > 0 {
		m.logScrollY--
	}
	return nil
}

func (m *LogViewModel) HandleDown(model *Model) tea.Cmd {
	m.scrollBy(model, 1)
	return nil
}

func (m *LogViewModel) HandleGoToEnd(model *Model) tea.Cmd {
	m.scrollToBottom(model)
	return nil
}

func (m *LogViewModel) HandleGoToBeginning() tea.Cmd {
	m.scrollToTop()
	return nil
}

func (m *LogViewModel) HandleSearch() tea.Cmd {
	m.searchMode = true
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.logSearchMatches = nil
	m.currentSearchIdx = 0
	return nil
}

func (m *LogViewModel) HandleFilter() tea.Cmd {
	m.filterMode = true
	m.filterCursorPos = len(m.filterText)
	return nil
}

func (m *LogViewModel) HandleNextSearchResult(model *Model) tea.Cmd {
	if len(m.logSearchMatches) == 0 {
		return nil
	}
	m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.logSearchMatches)
	m.scrollToSearchMatch(model, m.logSearchMatches[m.currentSearchIdx])
	return nil
}

func (m *LogViewModel) HandlePrevSearchResult(model *Model) tea.Cmd {
	if len(m.logSearchMatches) == 0 {
		return nil
	}
	m.currentSearchIdx--
	if m.currentSearchIdx < 0 {
		m.currentSearchIdx = len(m.logSearchMatches) - 1
	}
	m.scrollToSearchMatch(model, m.logSearchMatches[m.currentSearchIdx])
	return nil
}

func (m *LogViewModel) performFilter() {
	m.filteredLogs = nil
	if m.filterText == "" {
		return
	}

	filterText := strings.ToLower(m.filterText)

	for _, line := range m.logs {
		lineToSearch := strings.ToLower(line)
		if strings.Contains(lineToSearch, filterText) {
			m.filteredLogs = append(m.filteredLogs, line)
		}
	}
	m.scrollToTop()
}

func (m *LogViewModel) LogLines(model *Model, lines []string) {
	m.logs = append(m.logs, lines...)
	// Keep only last 10000 lines to prevent unbounded memory growth
	if len(m.logs) > 10000 {
		m.logs = m.logs[len(m.logs)-10000:]
	}

	// Keep the active filter and search results in sync with the live stream.
	if (m.filterMode || m.filterActive) && m.filterText != "" {
		scrollY, scrollVisual := m.logScrollY, m.logScrollVisual
		m.performFilter()
		m.logScrollY, m.logScrollVisual = scrollY, scrollVisual
		if m.WrapText() {
			m.clampVisualScroll(model)
		} else {
			m.logScrollY = min(m.logScrollY, m.calculateMaxScroll(model))
		}
	} else {
		// Auto-scroll to bottom only when not filtering
		m.scrollToBottom(model)
	}
	if m.searchText != "" && !m.searchMode {
		m.logSearchMatches = m.findLogSearchMatches(m.displayLogs())
		if m.currentSearchIdx >= len(m.logSearchMatches) {
			m.currentSearchIdx = max(0, len(m.logSearchMatches)-1)
		}
	} else {
		m.logSearchMatches = nil
	}
}

func (m *LogViewModel) FilterDeleteLastChar() {
	updated := m.FilterViewModel.FilterDeleteLastChar()
	if updated {
		m.performFilter()
	}
}

func (m *LogViewModel) Title() string {
	title := fmt.Sprintf("Logs: %s", m.container.Title())

	if m.WrapText() {
		title += " [wrap]"
	} else {
		title += " [nowrap]"
	}
	if m.prettyJSON {
		title += " [JSON]"
	}

	// Add search or filter status to title
	if (m.filterMode || m.filterActive) && m.filterText != "" {
		filterCount := len(m.filteredLogs)
		title += fmt.Sprintf(" - Filter: '%s' (%d/%d lines)", m.filterText, filterCount, len(m.logs))
	} else if len(m.logSearchMatches) > 0 {
		var statusParts []string
		if m.searchIgnoreCase {
			statusParts = append(statusParts, "i")
		}
		if m.searchRegex {
			statusParts = append(statusParts, "r")
		}

		statusStr := ""
		if len(statusParts) > 0 {
			statusStr = fmt.Sprintf(" [%s]", strings.Join(statusParts, ""))
		}

		title += fmt.Sprintf(" - Search: %d/%d%s", m.currentSearchIdx+1, len(m.logSearchMatches), statusStr)
	} else if m.searchText != "" && !m.searchMode {
		title += " - No matches found"
	}

	return title
}

func (m *LogViewModel) ToggleJSONFormatting(model *Model) {
	m.prettyJSON = !m.prettyJSON
	if m.searchText != "" && !m.searchMode {
		m.logSearchMatches = m.findLogSearchMatches(m.displayLogs())
		if m.currentSearchIdx >= len(m.logSearchMatches) {
			m.currentSearchIdx = max(0, len(m.logSearchMatches)-1)
		}
		m.clampVisualScroll(model)
	}
}

func (m *LogViewModel) HandleCancel() tea.Cmd {
	m.stopLogReader()
	return nil
}

func (m *LogViewModel) HandlePageUp(model *Model) tea.Cmd {
	pageSize := model.PageSize() - 2
	if pageSize < 1 {
		pageSize = 1
	}
	m.scrollBy(model, -pageSize)
	return nil
}

func (m *LogViewModel) HandleToggleWrap() tea.Cmd {
	m.ToggleWrapText()
	m.logScrollX = 0
	m.logScrollVisual = 0
	return nil
}

func (m *LogViewModel) HandleScrollLeft() tea.Cmd {
	if m.WrapText() || m.logScrollX <= 0 {
		return nil
	}
	m.logScrollX--
	return nil
}

func (m *LogViewModel) HandleScrollRight(model *Model) tea.Cmd {
	if m.WrapText() {
		return nil
	}

	maxScrollX := m.calculateMaxScrollX(model, m.displayLogs())
	if m.logScrollX < maxScrollX {
		m.logScrollX++
	}
	return nil
}

func (m *LogViewModel) HandlePageDown(model *Model) tea.Cmd {
	pageSize := model.PageSize() - 2
	if pageSize < 1 {
		pageSize = 1
	}
	m.scrollBy(model, pageSize)
	return nil
}
