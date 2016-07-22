package templates_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airware/vili/templates"
)

const (
	testTemplate             templates.Template = `KEY1 = {VAR1}`
	testTemplateDefault      templates.Template = `KEY1 = {VAR1:default1}`
	testTemplateEmptyDefault templates.Template = `KEY1 = {VAR1:}`
)

var testVariables = map[string]string{
	"VAR1": "VALUE1",
}

func TestParsing(t *testing.T) {
	populated1, invalid1 := testTemplate.Populate(nil)
	assert.Equal(t, `KEY1 = `, string(populated1))
	assert.Equal(t, true, invalid1)
	populated2, invalid2 := testTemplate.Populate(testVariables)
	assert.Equal(t, `KEY1 = VALUE1`, string(populated2))
	assert.Equal(t, false, invalid2)
}

func TestParsingDefaults(t *testing.T) {
	populated1, invalid1 := testTemplateDefault.Populate(nil)
	assert.Equal(t, `KEY1 = default1`, string(populated1))
	assert.Equal(t, false, invalid1)
	populated2, invalid2 := testTemplateDefault.Populate(testVariables)
	assert.Equal(t, `KEY1 = VALUE1`, string(populated2))
	assert.Equal(t, false, invalid2)
}

func TestParsingEmptyDefaults(t *testing.T) {
	populated1, invalid1 := testTemplateEmptyDefault.Populate(nil)
	assert.Equal(t, `KEY1 = `, string(populated1))
	assert.Equal(t, false, invalid1)
	populated2, invalid2 := testTemplateDefault.Populate(testVariables)
	assert.Equal(t, `KEY1 = VALUE1`, string(populated2))
	assert.Equal(t, false, invalid2)
}
