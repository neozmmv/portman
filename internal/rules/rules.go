package rules

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Proto string

const (
	TCP Proto = "tcp"
	UDP Proto = "udp"

	BlockBegin = "#PORTMAN BEGIN"
	BlockEnd   = "#PORTMAN END"
)

func NormalizeProtos(p string) ([]Proto, error) {
	p = strings.ToLower(strings.TrimSpace(p))

	switch p {
	case "tcp":
		return []Proto{TCP}, nil
	case "udp":
		return []Proto{UDP}, nil
	case "tcp/udp", "udp/tcp":
		return []Proto{TCP, UDP}, nil
	default:
		return nil, fmt.Errorf("invalid proto %q", p)
	}
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d", port)
	}
	return nil
}

func RuleLine(port int, proto Proto) string {
	return fmt.Sprintf(`-A INPUT -p %s -m %s --dport %d -j ACCEPT`, proto, proto, port)
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

type filterInfo struct {
	filterIdx int
	commitIdx int
}

func findFilter(lines []string) (filterInfo, error) {
	inFilter := false
	filterIdx := -1

	for i, line := range lines {
		t := strings.TrimSpace(line)
		if t == "*filter" {
			inFilter = true
			filterIdx = i
			continue
		}
		if inFilter && t == "COMMIT" {
			return filterInfo{filterIdx: filterIdx, commitIdx: i}, nil
		}
	}

	if filterIdx == -1 {
		return filterInfo{}, errors.New("missing *filter table")
	}
	return filterInfo{}, errors.New("missing COMMIT inside *filter")
}

func findChainHeaderEnd(lines []string, filterIdx int) int {
	i := filterIdx + 1
	for i < len(lines) {
		if strings.HasPrefix(lines[i], ":") {
			i++
			continue
		}
		break
	}
	return i
}

func ensurePortmanBlockAtTop(lines []string, fi filterInfo) ([]string, int, int, bool, error) {
	beginIdx := -1
	endIdx := -1

	for i := fi.filterIdx + 1; i < fi.commitIdx; i++ {
		if lines[i] == BlockBegin {
			beginIdx = i
			continue
		}
		if lines[i] == BlockEnd {
			endIdx = i
			break
		}
	}

	if beginIdx != -1 && endIdx != -1 && beginIdx < endIdx {
		return lines, beginIdx, endIdx, false, nil
	}

	if beginIdx != -1 || endIdx != -1 {
		return nil, -1, -1, false, errors.New("invalid PORTMAN block: only BEGIN or END found")
	}

	insertAt := findChainHeaderEnd(lines, fi.filterIdx)

	out := make([]string, 0, len(lines)+2)
	out = append(out, lines[:insertAt]...)
	out = append(out, BlockBegin)
	out = append(out, BlockEnd)
	out = append(out, lines[insertAt:]...)

	return out, insertAt, insertAt + 1, true, nil
}

func Open(content string, port int, protoInput string) (string, bool, error) {
	if err := ValidatePort(port); err != nil {
		return "", false, err
	}

	protos, err := NormalizeProtos(protoInput)
	if err != nil {
		return "", false, err
	}

	content = normalizeNewlines(content)
	lines := strings.Split(content, "\n")

	fi, err := findFilter(lines)
	if err != nil {
		return "", false, err
	}

	lines, beginIdx, endIdx, created, err := ensurePortmanBlockAtTop(lines, fi)
	_ = beginIdx
	if err != nil {
		return "", false, err
	}

	if created {
		fi, err = findFilter(lines)
		if err != nil {
			return "", false, err
		}
		for i := fi.filterIdx + 1; i < fi.commitIdx; i++ {
			if lines[i] == BlockEnd {
				endIdx = i
				break
			}
		}
	}

	existing := make(map[string]struct{}, len(lines))
	for _, l := range lines {
		existing[l] = struct{}{}
	}

	toInsert := make([]string, 0, len(protos))
	for _, pr := range protos {
		line := RuleLine(port, pr)
		if _, ok := existing[line]; !ok {
			toInsert = append(toInsert, line)
		}
	}

	if len(toInsert) == 0 && !created {
		return strings.Join(lines, "\n"), false, nil
	}

	out := make([]string, 0, len(lines)+len(toInsert))
	out = append(out, lines[:endIdx]...)
	out = append(out, toInsert...)
	out = append(out, lines[endIdx:]...)

	return strings.Join(out, "\n"), true, nil
}

func Close(content string, port int, protoInput string) (string, bool, error) {
	if err := ValidatePort(port); err != nil {
		return "", false, err
	}

	protos, err := NormalizeProtos(protoInput)
	if err != nil {
		return "", false, err
	}

	content = normalizeNewlines(content)
	lines := strings.Split(content, "\n")

	fi, err := findFilter(lines)
	if err != nil {
		return "", false, err
	}

	beginIdx := -1
	endIdx := -1
	for i := fi.filterIdx + 1; i < fi.commitIdx; i++ {
		if lines[i] == BlockBegin {
			beginIdx = i
			continue
		}
		if lines[i] == BlockEnd {
			endIdx = i
			break
		}
	}

	if beginIdx == -1 || endIdx == -1 || beginIdx >= endIdx {
		return strings.Join(lines, "\n"), false, nil
	}

	targets := make(map[string]struct{}, len(protos))
	for _, pr := range protos {
		targets[RuleLine(port, pr)] = struct{}{}
	}

	changed := false
	out := make([]string, 0, len(lines))

	for i, l := range lines {
		if i > beginIdx && i < endIdx {
			if _, ok := targets[l]; ok {
				changed = true
				continue
			}
		}
		out = append(out, l)
	}

	return strings.Join(out, "\n"), changed, nil
}

func Status(content string, port int, protoInput string) (map[Proto]bool, error) {
	if err := ValidatePort(port); err != nil {
		return nil, err
	}

	protos, err := NormalizeProtos(protoInput)
	if err != nil {
		return nil, err
	}

	content = normalizeNewlines(content)
	lines := strings.Split(content, "\n")

	existing := make(map[string]struct{}, len(lines))
	for _, l := range lines {
		existing[l] = struct{}{}
	}

	res := make(map[Proto]bool, len(protos))
	for _, pr := range protos {
		line := RuleLine(port, pr)
		_, ok := existing[line]
		res[pr] = ok
	}

	return res, nil
}

type PortRule struct {
	Port  int
	Proto Proto
}

func parsePortmanAcceptRule(line string) (PortRule, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return PortRule{}, false
	}
	if strings.HasPrefix(line, "#") {
		return PortRule{}, false
	}

	fields := strings.Fields(line)
	if len(fields) < 8 {
		return PortRule{}, false
	}
	// Expect something like:
	// -A INPUT -p tcp -m tcp --dport 443 -j ACCEPT
	if fields[0] != "-A" || fields[1] != "INPUT" {
		return PortRule{}, false
	}

	var proto Proto
	port := 0
	jumpAccept := false

	for i := 2; i < len(fields); i++ {
		switch fields[i] {
		case "-p":
			if i+1 >= len(fields) {
				return PortRule{}, false
			}
			p := strings.ToLower(fields[i+1])
			switch p {
			case string(TCP):
				proto = TCP
			case string(UDP):
				proto = UDP
			default:
				return PortRule{}, false
			}
			i++
		case "--dport":
			if i+1 >= len(fields) {
				return PortRule{}, false
			}
			p, err := strconv.Atoi(fields[i+1])
			if err != nil {
				return PortRule{}, false
			}
			port = p
			i++
		case "-j":
			if i+1 >= len(fields) {
				return PortRule{}, false
			}
			jumpAccept = strings.ToUpper(fields[i+1]) == "ACCEPT"
			i++
		}
	}

	if proto == "" || port == 0 || !jumpAccept {
		return PortRule{}, false
	}
	if err := ValidatePort(port); err != nil {
		return PortRule{}, false
	}

	return PortRule{Port: port, Proto: proto}, true
}

// List returns all open ports managed inside the #PORTMAN block.
//
// It only includes rules inside the *filter table between #PORTMAN BEGIN/END,
// and only rules of the form "-A INPUT ... --dport <port> -j ACCEPT".
func List(content string) ([]PortRule, error) {
	content = normalizeNewlines(content)
	lines := strings.Split(content, "\n")

	fi, err := findFilter(lines)
	if err != nil {
		return nil, err
	}

	beginIdx := -1
	endIdx := -1
	for i := fi.filterIdx + 1; i < fi.commitIdx; i++ {
		if lines[i] == BlockBegin {
			beginIdx = i
			continue
		}
		if lines[i] == BlockEnd {
			endIdx = i
			break
		}
	}

	if beginIdx == -1 || endIdx == -1 || beginIdx >= endIdx {
		return []PortRule{}, nil
	}

	seen := make(map[PortRule]struct{})
	res := make([]PortRule, 0)

	for i := beginIdx + 1; i < endIdx; i++ {
		r, ok := parsePortmanAcceptRule(lines[i])
		if !ok {
			continue
		}
		if _, exists := seen[r]; exists {
			continue
		}
		seen[r] = struct{}{}
		res = append(res, r)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Port != res[j].Port {
			return res[i].Port < res[j].Port
		}
		return res[i].Proto < res[j].Proto
	})

	return res, nil
}
