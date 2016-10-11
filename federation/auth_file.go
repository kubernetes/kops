package federation

import (
	"strings"
	"fmt"
	"bytes"
)

type AuthFile struct {
	Lines []*AuthFileLine
}

type AuthFileLine struct {
	User   string
	Secret string
	Role   string
}

func ParseAuthFile(data []byte) (*AuthFile, error) {
	parsed := &AuthFile{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parsedLine, err := ParseAuthFileLine(line)
		if err != nil {
			return nil, err
		}
		parsed.Lines = append(parsed.Lines, parsedLine)
	}
	return parsed, nil
}

func (a*AuthFile) FindUser(user string) *AuthFileLine {
	for _, line := range a.Lines {
		if line.User == user {
			return line
		}
	}
	return nil
}

func (a*AuthFile) Add(line *AuthFileLine) error {
	existing := a.FindUser(line.User)
	if existing != nil {
		return fmt.Errorf("user %q already exists in file", line.User)
	}
	a.Lines = append(a.Lines, line)
	return nil
}

func (a*AuthFile) Encode() string {
	var b bytes.Buffer

	for _, line := range a.Lines {
		b.WriteString(fmt.Sprintf("%s,%s,%s\n", line.Secret, line.User, line.Role))
	}
	return b.String()
}

func ParseAuthFileLine(line string) (*AuthFileLine, error) {
	tokens := strings.Split(line, ",")
	if len(tokens) != 3 {
		return nil, fmt.Errorf("unexpected line: expected exactly 3 tokens, found %d", len(tokens))
	}
	parsed := &AuthFileLine{
		Secret: tokens[0],
		User: tokens[1],
		Role: tokens[2],
	}
	return parsed, nil
}