package enrich

import (
	"testing"

	"github.com/sirupsen/logrus"
)

type Blank struct {
}

func (*Blank) Enrich(entry *logrus.Entry) *logrus.Entry {
	return entry.WithField("addendum", "put your object specific info here")
}

type Body struct {
	Front  string
	Middle string
	Tail   string
}

func (b *Body) Enrich(entry *logrus.Entry) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"front":  b.Front,
		"center": b.Middle,
		"back":   b.Tail})
}

func TestEnrichment(t *testing.T) {
	logger := logrus.New()
	entry := logrus.NewEntry(logger).WithField("parentKey", "parentValue")
	entry.Info("test")

	blank := &Blank{}
	bentry := Enrich(entry, blank)
	bentry.Info("test2")

	body := &Body{
		Front:  "11111111111111111111111111",
		Middle: "22222222222222222222222222",
		Tail:   "33333333333333333333333333",
	}
	bentry = Enrich(entry, body)
	bentry.Info("test3")
}
