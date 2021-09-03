package weblinks

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/hashicorp/logutils"
)

// Links is a mapping of rel strings to Link objects
type Links map[string]*Link

// Link is the link URI and associated attributes
type Link struct {
	URI        *url.URL
	Attributes map[string]string
	rels       []string
}

type params struct {
	rels       []string
	attributes map[string]string
}

type param struct {
	key   string
	value string
}

func init() {
	var minLevel = ""
	if os.Getenv("WEBLINKS_DEBUG") != "" {
		minLevel = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG"},
		MinLevel: logutils.LogLevel(minLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
}

func parseLink(s string) (*Link, string, error) {
	if s[0] != '<' {
		return nil, s, fmt.Errorf("link does not begin with <")
	}

	s = s[1:] // consume <

	endUriIndex := strings.Index(s, ">")
	if endUriIndex == -1 {
		return nil, s, fmt.Errorf("link does not end with >")
	}

	uri, err := url.Parse(s[0:endUriIndex])

	if err != nil {
		return nil, s, fmt.Errorf("link could not be parsed: %w", err)
	}

	result := Link{
		URI: uri,
	}

	s = strings.TrimSpace(s[endUriIndex+1:])

	if len(s) == 0 || s[0] != ';' {
		return &result, s, fmt.Errorf("no params found")
	}

	s = s[1:] // consume ;

	params, rest, err := parseParams(s)

	if err != nil {
		return &result, s, fmt.Errorf("params could not be parsed: %w", err)
	}

	result.rels = params.rels
	result.Attributes = params.attributes

	// Allow "anchor" param to override parsed URI
	anchor, hasAnchor := params.attributes["anchor"]
	if hasAnchor {
		anchorUri, err := url.Parse(anchor)

		if err == nil {
			result.URI = anchorUri
		}
	}

	return &result, rest, nil
}

func parseParams(s string) (*params, string, error) {
	// Examples
	//  rel=\"previous\"; title=\"hello, God; it's me, margaret\"

	result := params{
		rels:       make([]string, 0, 1),
		attributes: make(map[string]string),
	}

	for {
		if len(s) == 0 {
			break
		}

		// Signals next link
		if s[0] == ',' {
			s = s[1:] // consume ","
			s = strings.TrimSpace(s)
			return &result, s, nil
		} else if s[0] == ';' {
			s = s[1:] // consume ";"
			s = strings.TrimSpace(s)
			continue
		}

		p, rest, err := parseParam(s)

		if err != nil {
			return &result, s, err
		}

		switch p.key {
		case "rel":
			result.rels = strings.Split(p.value, " ")
		default:
			result.attributes[p.key] = p.value
		}

		s = strings.TrimSpace(rest)
	}
	return &result, s, nil
}

func parseParam(s string) (*param, string, error) {
	s = strings.TrimSpace(s)
	key, s, err := parseToken(s)

	if err != nil {
		return nil, s, err
	}

	if len(s) == 0 {
		return nil, s, fmt.Errorf("expected '=' but found end of string")
	}

	if s[0] != '=' {
		return nil, s, fmt.Errorf("expected '=' but found '%v'", (string)(s[0]))
	}

	s = s[1:] // consume "="

	value, rest, err := parseParamValue(s)

	if err != nil {
		return nil, s, fmt.Errorf("could not parse param value: %w", err)
	}

	log.Printf("[DEBUG] parseParamValue %v, remaining = \"%v\"", value, rest)

	return &param{
		key:   key,
		value: value,
	}, rest, nil
}

func parseParamValue(s string) (string, string, error) {
	var value string
	var rest string

	if strings.HasPrefix(s, "\"") {
		// Quoted string, look for end of quote
		endValueIndex := strings.Index(s[1:], "\"")
		if endValueIndex == -1 {
			return "", "", fmt.Errorf("expected \" but found none")
		}
		value = s[1 : endValueIndex+1]
		rest = s[endValueIndex+1:]

		rest = rest[1:] // consume "
	} else {
		// Nonquoted string, just look for either a space, comma, or semi
		endValueIndex := strings.IndexAny(s, " ;,")
		if endValueIndex == -1 {
			value = s[0:]
			rest = ""
		} else {
			value = s[0:endValueIndex]
			rest = s[endValueIndex+1:]
		}

		log.Printf("[DEBUG] nonquoted: value = \"%v\", rest = \"%v\"", value, rest)
	}

	return value, rest, nil
}

func parseToken(s string) (string, string, error) {
	endToken := strings.IndexFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	if endToken == -1 {
		return "", s, fmt.Errorf("no token found")
	}

	log.Printf("[DEBUG] parseToken = %v", s[0:endToken])

	token := s[0:endToken]
	s = s[endToken:]
	return token, s, nil
}

// Parse parses a set of web links, usually delivered by Link header,
// according to [RFC5988](https://datatracker.ietf.org/doc/html/rfc5988)
func Parse(s string) (Links, error) {
	result := make(Links)

	s = strings.TrimSpace(s)

	for {
		if s == "" {
			break
		}

		link, rest, err := parseLink(s)

		log.Printf("[DEBUG] after parseLink rest = \"%v\"", rest)

		if err != nil {
			return result, err
		}

		// links can be associated with multiple rels, separated by a space (section 5)
		for _, r := range link.rels {
			result[r] = link
		}

		s = strings.TrimSpace(rest)
	}

	return result, nil
}
