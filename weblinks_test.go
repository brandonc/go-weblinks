package weblinks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeblinks_Parse(t *testing.T) {
	t.Run("returns empty map for empty string", func(t *testing.T) {
		links, err := Parse("")
		assert.Nil(t, err)
		assert.Equal(t, 0, len(links))
	})

	t.Run("parses simple syntax", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter2>; rel=\"previous\"")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(links))

		previous, ok := links["previous"]
		assert.True(t, ok)
		assert.NotNil(t, previous)
		assert.Equal(t, "http://example.com/TheBook/chapter2", previous.URI.String())
	})

	t.Run("parses multiple params", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter2>; rel=\"previous\"; title=\"hello world\"")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(links))

		previous, ok := links["previous"]
		assert.True(t, ok)
		assert.NotNil(t, previous)
		assert.Equal(t, "http://example.com/TheBook/chapter2", previous.URI.String())
		assert.Equal(t, "hello world", previous.Attributes["title"])
	})

	t.Run("parses multiple links", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter2>; rel=\"previous\", <http://example.com/TheBook/chapter4>; rel=\"next\", <http://example.com/TheBook/chapter31>; rel=\"last\", <http://example.com/TheBook/chapter1>; rel=\"first\"")

		assert.Nil(t, err)
		assert.Equal(t, 4, len(links))

		previous, ok := links["previous"]
		assert.True(t, ok)
		assert.NotNil(t, previous)
		assert.Equal(t, "http://example.com/TheBook/chapter2", previous.URI.String())

		next, ok := links["next"]
		assert.True(t, ok)
		assert.NotNil(t, next)
		assert.Equal(t, "http://example.com/TheBook/chapter4", next.URI.String())

		first, ok := links["first"]
		assert.True(t, ok)
		assert.NotNil(t, first)
		assert.Equal(t, "http://example.com/TheBook/chapter1", first.URI.String())

		last, ok := links["last"]
		assert.True(t, ok)
		assert.NotNil(t, last)
		assert.Equal(t, "http://example.com/TheBook/chapter31", last.URI.String())
	})

	t.Run("commas and semis within quoted strings", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter2>; rel=\"previous\"; title=\"hello, God; it's me, margaret\"")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(links))

		previous, ok := links["previous"]
		assert.True(t, ok)
		assert.NotNil(t, previous)
		assert.Equal(t, "http://example.com/TheBook/chapter2", previous.URI.String())
		assert.Equal(t, "hello, God; it's me, margaret", previous.Attributes["title"])
	})

	t.Run("when semicolon missing", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter2>; rel=\"previous\" title=\"Chapter Two\"")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(links))

		previous, ok := links["previous"]
		assert.True(t, ok)

		assert.Equal(t, "Chapter Two", previous.Attributes["title"])
	})

	t.Run("overrides URI with anchor param", func(t *testing.T) {
		links, err := Parse("<>; rel=\"last\"; anchor=\"http://example.com/TheBook/chapter31/\"")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(links))

		last, ok := links["last"]
		assert.True(t, ok)
		assert.NotNil(t, last)

		assert.Equal(t, "http://example.com/TheBook/chapter31/", last.URI.String())
	})

	t.Run("parses multiple rels", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter31>; rel=\"last http://example.com/rel-ext/bonus\";")

		assert.Nil(t, err)
		assert.Equal(t, 2, len(links))

		last, ok := links["last"]
		assert.True(t, ok)
		assert.NotNil(t, last)
		assert.Equal(t, last.URI.String(), "http://example.com/TheBook/chapter31")

		bonus, ok := links["http://example.com/rel-ext/bonus"]
		assert.True(t, ok)
		assert.NotNil(t, bonus)
		assert.Equal(t, bonus.URI.String(), "http://example.com/TheBook/chapter31")
	})

	t.Run("parses target attributes", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter31>; rel=\"last\"; type=\"text/html\"; media=\"screen\"; hreflang=en/us")

		assert.Nil(t, err)
		last, ok := links["last"]
		assert.True(t, ok)
		assert.NotNil(t, last)

		assert.Equal(t, "text/html", last.Attributes["type"])
		assert.Equal(t, "screen", last.Attributes["media"])
		assert.Equal(t, "en/us", last.Attributes["hreflang"])
		assert.Equal(t, last.URI.String(), "http://example.com/TheBook/chapter31")
	})

	t.Run("parses nonquoted params", func(t *testing.T) {
		links, err := Parse("<http://example.com/TheBook/chapter31>; rel=\"last\"; type=text/html; media=screen ; hreflang=en/us")

		assert.Nil(t, err)
		last, ok := links["last"]
		assert.True(t, ok)
		assert.NotNil(t, last)

		assert.Equal(t, "text/html", last.Attributes["type"])
		assert.Equal(t, "screen", last.Attributes["media"])
		assert.Equal(t, "en/us", last.Attributes["hreflang"])
		assert.Equal(t, last.URI.String(), "http://example.com/TheBook/chapter31")
	})

	t.Run("does not parse title* attribute", func(t *testing.T) {
		_, err := Parse("<http://example.com/TheBook/chapter31>; rel=\"last\"; title*=UTF-8'de'n%c3%a4chstes%20Kapitel, <http://example.com/TheBook/chapter1>; rel=\"first\"; title*=us-ascii'en-us'hello%20world")

		assert.NotNil(t, err)
		assert.Equal(t, "params could not be parsed: expected '=' but found '*'", err.Error())
	})
}
