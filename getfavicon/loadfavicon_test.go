// Copyright @lolorenzo777 - 2022 May

package getfavicon

import (
	"log"
	"net/http"
	"os"

	//	"os"
	"testing"
)

var gwebsitesOK = []string{
	"https://go.dev/",
	"https://brave.com/",
	"https://github.com/",
	"https://twitter.com/",
	"https://www.linkedin.com/",
	"https://protonmail.com/",
	"https://getbootstrap.com/",
	"https://www.cloudflare.com/",
	"https://www.docker.com/",
}

var gwebsitesKO = []string{
	"email://www.dummy.com",
	"http:dummy.io",
	"www.dummy.abc",
	"github.com",
}

var fNeedCleaning bool

const (
	testDIR = "downloadedicons"
)

func init() {
	os.RemoveAll(testDIR)
}

func TestSlugName(t *testing.T) {
	str := SlugHost("https://go.dev/")
	if str != "go-dev" {
		t.Error("https://go.dev/ --> " + str)
	}

	str = SlugHost("https://lolorenzo777.github.io/website4tests-1") 
	if str != "lolorenzo777-github-iowebsite4tests-1" {
		t.Error("https://lolorenzo777.github.io/website4tests-1" + str)
	}
}

func TestGetFaviconLinks(t *testing.T) {
    // Create the HTTP client, re-usable, with timeout
    client := &http.Client{}
	for _, v := range(gwebsitesOK) {
		log.Printf("---getfaviconLinks: %s\n", v)
		if _, err := getFaviconLinks(client, v); err != nil {
			t.Error(err)
		}
	}

	for _, v := range(gwebsitesKO) {
		if _, err := getFaviconLinks(client, v); err == nil {
			t.Errorf("---getfaviconLinks: %s\n", v)
		}
	}
}

func TestGetfavicons(t *testing.T) {
	Website := "https://github.com/"
	favicons, err := ReadAll(Website);
	if  err != nil {
		t.Errorf("---getFavicons %q: %v\n", Website, err)
	}
	if len(favicons) == 0 {
		t.Fail()
	}
}

func TestDownloadFaviconsDummyWebsite(t *testing.T) {
	favs, err := Download("https://www.dummy.dummy", ".test/"+testDIR, false)
	if err.Error()[len(err.Error())-12:] != "no such host" {
		t.Error(err)
	}
	if len(favs) != 0 {
		t.Fail()
	}
}

func TestDownloadFaviconsNone(t *testing.T) {
	favs, err := Download("https://lolorenzo777.github.io/website4tests-1", ".test/"+testDIR, false)
	if err != nil {
		t.Error(err)
	}
	if len(favs) != 0 {
		t.Fail()
		fNeedCleaning = true
	}
}

func TestDownloadFaviconsSingle(t *testing.T) {
	favs, err := Download("https://github.com/", ".test/"+testDIR, true)
	if err != nil {
		t.Error(err)
	}
	if len(favs) != 1 {
		t.Fail()
	}

	favs, err = Download("https://lolorenzo777.github.io/website4tests-2/", ".test/"+testDIR, true)
	if err != nil {
		t.Error(err)
	}
	if len(favs) != 1 || favs[0].DiskFileName != "lolorenzo777-github-iowebsite4tests-2+test-32x32.png" {
		t.Fail()
	}

	fNeedCleaning = true
}


func TestDownloadFaviconsMultiple(t *testing.T) {
	favs, err := Download("https://www.docker.com", ".test/"+testDIR, false)
	if err != nil {
		t.Error(err)
	}
	if len(favs) <= 1 {
		t.Fail()
	}
	fNeedCleaning = true
}

func TestDownloadFaviconsBatch(t *testing.T) {
	for _, v := range(gwebsitesOK) {
		log.Printf("---Download favicon from %q\n", v)
		favs, err := Download(v, ".test/"+testDIR, false)
		if err != nil {
			t.Error(err)
		}
		if len(favs) == 0 {
			t.Fail() 
		}
	}
	fNeedCleaning = true
}

func TestClear(t *testing.T) {
	if fNeedCleaning {
		os.RemoveAll(".test")
	}
}