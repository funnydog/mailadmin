package urls

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestURLManager(t *testing.T) {
	router := httprouter.New()

	manager := CreateManager(router)

	myURL := URL{
		Prefix:      "/",
		Method:      "PUT",
		HandlerFunc: func(http.ResponseWriter, *http.Request) {},
		Name:        "friendly-name",
	}

	err := manager.Add(&myURL)
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	myURL.Method = "POST"
	err = manager.Add(&myURL)
	if err != nil {
		t.Error(err)
		return
	}

	err = manager.Add(&myURL)
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	myURL.Name = "other-name"
	err = manager.Add(&myURL)
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	myURL.Method = "GET"
	myURL.Name = ""
	err = manager.Add(&myURL)
	if err != nil {
		t.Error(err)
		return
	}

	reverse, err := manager.Reverse("friendly-name", nil)
	if err != nil {
		t.Error(err)
		return
	}

	if reverse != myURL.Prefix {
		t.Errorf("Expected '%s' but got '%s' instead", myURL.Prefix, reverse)
	}
}
