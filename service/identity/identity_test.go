package identity

import (
	"reflect"
	"testing"

	"github.com/arwoosa/notifaction/service"
)

func TestClassificationLang_Add(t *testing.T) {
	// Test case 1: Add a new language and info
	cl := newClassficationLang()
	info1 := &service.Info{Sub: "sub1", Name: "name1", Email: "email1", Enable: true}
	cl.add("lang1", info1)
	if !reflect.DeepEqual(cl.keys, []string{"lang1"}) {
		t.Errorf("Expected keys to be [lang1], got %v", cl.keys)
	}
	if !reflect.DeepEqual(cl.data["lang1"], []*service.Info{info1}) {
		t.Errorf("Expected data[lang1] to be [%v], got %v", []*service.Info{info1}, cl.data["lang1"])
	}

	// Test case 2: Add another info to an existing language
	info2 := &service.Info{Sub: "sub2", Name: "name2", Email: "email2", Enable: false}
	cl.add("lang1", info2)
	if !reflect.DeepEqual(cl.keys, []string{"lang1"}) {
		t.Errorf("Expected keys to be [lang1], got %v", cl.keys)
	}
	if !reflect.DeepEqual(cl.data["lang1"], []*service.Info{info1, info2}) {
		t.Errorf("Expected data[lang1] to be [%v, %v], got %v", info1, info2, cl.data["lang1"])
	}

	// Test case 3: Add a new language and info to an empty classificationLang
	cl = newClassficationLang()
	info3 := &service.Info{Sub: "sub3", Name: "name3", Email: "email3", Enable: true}
	cl.add("lang3", info3)
	if !reflect.DeepEqual(cl.keys, []string{"lang3"}) {
		t.Errorf("Expected keys to be [lang3], got %v", cl.keys)
	}
	if !reflect.DeepEqual(cl.data["lang3"], []*service.Info{info3}) {
		t.Errorf("Expected data[lang3] to be [%v], got %v", []*service.Info{info3}, cl.data["lang3"])
	}
}
