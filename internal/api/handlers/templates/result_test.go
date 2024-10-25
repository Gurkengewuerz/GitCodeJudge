package templates_test

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers/templates"
	"testing"
)

func TestGetResultTemplate(t *testing.T) {
	tmpl := templates.GetResultTemplate()
	if tmpl == nil {
		t.Fatal("Expected template to be created")
	}

	expected := "results"
	if tmpl.Name() != expected {
		t.Fatalf("Expected template name to be %s, got %s", expected, tmpl.Name())
	}
}
