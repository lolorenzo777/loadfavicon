// Copyright @lolorenzo777 - 2022 May

package getfavicon

import (
	"bytes"
    "io"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosimple/slug"
)

// lookuprel contains the list of rel id that are usally related to
// define the favicon file/url into the <Head> section of the website's webpage.
// <link rel="{lookuprel}"...>
var lookuprel = []string{
    "icon",
    "shortcut icon",
    "apple-touch-icon",
    "apple-touch-icon-precomposed",
    "mask-icon",
    }

// acceptedMIMEtypes contains the list of accepted MIME types for content of icon files
// https://developer.mozilla.org/fr/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
var acceptedMIMEtypes = []string{
    "image/x-icon",
    "image/png",
    "image/svg+xml",
    "image/jpeg",
    "image/webp",
}

var validIconFileExt = []string{
    ".ico",
    ".png",
    ".svg",
    ".jpg",
    ".jpeg",
}

type TFavicon struct {
    Website url.URL // The absolute URL of the favicon's host website
    Webicon url.URL // The absolute URL of the favicon's file
    DiskFileName string // The disk file name, based on the slugyfied website URL and the favicon url name
    Color string // Color specfications if any specified in the <link> node
    Size string // Size specfications if any specified in the <link> node
    Image []byte // The loaded raw image
}



// find is an helper to look for a specific item in a slice.
//
// Returns the index of the value found, and -1 if value not found
func find(list []string, value string) int {
    for i, v := range(list) {
        if v == value {
            return i
        }
    }
    return -1
}

// parseURL builds an absolute and valid URL to look for an icon. 
// http schema is added if missing. user, rawQuery and Fragments are cleaned-up if any. 
// website can describe an ansolute or a relative path.
// 
// Set clearfile to remove any filename at the end of the path
//
// Returns nil if not http not https schema, if unable to parse website, or if host is different from the one defined in website
func parseURL(host *url.URL, website string, clearfile bool) *url.URL {
    url, err := url.Parse(website)
    if err != nil {
        log.Println(err)
        return nil
    }
    if len(url.Host) == 0 {
        if host != nil {
            url.Host = host.Host
            url.Path = filepath.Join(host.Path, url.Path)
        } else {
            return nil
        }
    } 
    if len(url.Scheme) == 0 {
        url.Scheme = "http"
    } else if url.Scheme != "http" && url.Scheme != "https"  {
        return nil
    }
    url.User = nil
    url.RawQuery = ""
    url.Fragment = ""
    if clearfile {
        url.Path = strings.TrimPrefix(url.Path, "/")
    }
    
    return url
}

// doHttpGETRequest create, setup, and send a http GET request.
// 
// Returns the hhtp response. client is not closed and can be reused.
func doHttpGETRequest(client *http.Client, getrequest string) (*http.Response, error) {
    // create and setup http request before sending
    req, err := http.NewRequest("GET", getrequest, nil)
    if err != nil {
        return nil, err
    }
    // make the request
    req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
    //HACK: req.Header.Add("Accept-Encoding", "gzip, deflate, br")
    req.Header.Add("Accept-Language","fr-FR,fr,en-US,en")
    req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
    req.Header.Add("Sec-Fetch-Dest", "document")
    req.Header.Add("sec-fetch-mode", "navigate")
    req.Header.Add("sec-fetch-site", "none")
    req.Header.Add("Sec-Fetch-User", "?1")
    req.Header.Add("Upgrade-Insecure-Requests", "1")
    req.Header.Add("Referrer-Policy", "strict-origin-when-cross-origin")

    // send the request
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    return resp, nil
}

// SlugHost slugify the website URL 
func SlugHost(website string) string {
    hosturl := parseURL(nil, website, true)
    if hosturl == nil {
        return ""
    }
    return slug.Make(hosturl.Hostname() + hosturl.Path) 
}

// getFaviconLinks returns a list of urls for the favicons of this website.
// 
// It extracts the list of links declared in the <head> section of the site
// that may corresond to favicon, and add favicon.ico as a valid link to the list if the website has responded. 
func getFaviconLinks(client *http.Client, website string) (favicons []TFavicon, err error) {
    // ensure hosturl is a host url
    hosturl := parseURL(nil, website, true)
    if hosturl == nil {
        return nil, fmt.Errorf("fail to parse website url %q", website)
    }

    // add the favicon.ico to the list as the default file to lookup
    ico := TFavicon{
        Website: *hosturl,
        Webicon: *hosturl,
        DiskFileName: SlugHost(website) + "+favicon.ico"}
    ico.Webicon.Path += "favicon.ico"
    favicons = append(favicons, ico)
    
    // sending the request to the website
    req := hosturl.String()
    resp, err := doHttpGETRequest(client, req)
    if err != nil {
        return favicons, err
    }
    if resp.StatusCode != http.StatusOK {
        // if unable to reach the website
        // if there's a path, try the roothost if any
        if len(hosturl.Path) > 0 {
            resp.Body.Close()
            hosturl.Path = ""

            ico := TFavicon{
                Website: *hosturl,
                Webicon: *hosturl,
                DiskFileName: SlugHost(website) + "+favicon.ico"}
            ico.Webicon.Path += "favicon.ico"
            favicons[0] = ico

            req := hosturl.String()
            resp, err = doHttpGETRequest(client, req)
            if err != nil {
                return favicons, err
            }
        }
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return favicons, fmt.Errorf("unable to reach %q. Status code:%v", req, resp.Status)
    }
   
    // Create a goquery document from the HTTP response
    document, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        log.Println(err)
        return favicons, err
    }

    // build the list of favicon url related to this website
    document.Find("link").Each(func(i int, s *goquery.Selection){
        rel, hasrel := s.Attr("rel")
        href, hasref := s.Attr("href")
        if hasrel && hasref {
            rel = strings.ToLower(strings.Trim(rel, " "))
            if find(lookuprel, rel) >= 0 {
                if phref := parseURL(hosturl, href, false); phref != nil {
                    if find(validIconFileExt, filepath.Ext(phref.Path)) == -1 {
                        // ignore favicon files without a valid file extension
                        return
                    }
                    size, _ := s.Attr("sizes")
                    color, _ := s.Attr("color")
                    filename := SlugHost(website) + "+" + filepath.Base(phref.Path)
                    favicons = append(favicons, TFavicon{
                                                    Website: *hosturl,
                                                    Webicon: *phref,
                                                    DiskFileName: filename, 
                                                    Size:size, 
                                                    Color:color})    
                }
            }
        }
    })
  
    return favicons, nil
}

// ReadAll gets all favicons of a single website in memory. 
//
// Favicon's urls returned by getFaviconLinks are scanned. Only files 
// correspnding to valid image MIME formats (defined in var acceptedMIMEtypes)
// are returned. Duplicates are ignored.
func ReadAll(website string) (favicons []TFavicon, err error) {

    // Create the HTTP client, re-usable, with timeout
    client := &http.Client{Timeout:time.Second*5}

    faviconslinks, err := getFaviconLinks(client, website)
    if err != nil {
        return nil, err
    } 

    // scan all favicon links 
    for _, favicon := range(faviconslinks) {
        fnext := false
        faviconurl := favicon.Webicon.String()
        resp, err := doHttpGETRequest(client, faviconurl)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            // ignore unreadable files, for whatever reasons
            continue
        }
        // copy data from HTTP response to file
        b, _ := io.ReadAll(resp.Body)
        if len(b) == 0 {
            log.Printf("unable to get the content of the icon located at %q\n", faviconurl)
            continue
        }
        favicon.Image = b
        // avoid duplicate
        for _, existing := range(favicons) {
            if bytes.Equal(existing.Image, favicon.Image) {
                fnext = true
                continue
            }
        }
        if fnext {
            continue
        }
        // check content type
        contenttype := http.DetectContentType(favicon.Image)
        if find(acceptedMIMEtypes, contenttype) == -1 {
            // warning: DetectContentType detect SVG as text (see https://mimesniff.spec.whatwg.org/#identifying-a-resource-with-an-unknown-mime-type )
            if filepath.Ext(favicon.Webicon.Path) != ".svg" || !isValidSVG(favicon.Image) {
                // not an icon image ?!
                continue
            }           
        }
        favicons = append(favicons, favicon)
    }
    return favicons, nil
}

// SelectSingle selects a single favicon from favicons based on a simple rule.
// It selects .svg if any or selects the bigest size one if multiples one exists, finaly get the .ico if it exists
// Call LoadAll favicons before to build []TFavicon
//
// Returns nil if favicons was empty
func SelectSingle(favicons []TFavicon) (single *TFavicon) {
    // look for svg
    for _, one := range(favicons) {
        if filepath.Ext(one.DiskFileName) == ".svg" {
            single = new(TFavicon)
            *single = one
            return single
        }
    }
    // loop to look for bigest size or the ico file
    biggestSize := 0
    for _, one := range(favicons) {
        if biggestSize == 0 && filepath.Ext(one.DiskFileName) == ".ico" {
            single = new(TFavicon)
            *single = one
            continue
        }

        reader := bytes.NewReader(one.Image)
        cfg, _, err := image.DecodeConfig(reader);
        if err == nil {
            if biggestSize == 0 {
                biggestSize = cfg.Height * cfg.Width
                single = new(TFavicon)
                *single = one
            } else if cfg.Height * cfg.Width > biggestSize {
                biggestSize = cfg.Height * cfg.Width
                single = new(TFavicon)
                *single = one
            }
        }
    }
    return single
}

// Download loads all favicons files related to a website and store them locally to the 'toDir' directoty.
// Files are saved with name prefixed by savePrefix. Existing dest file are replaced.
// 'toDir' parameter can't be an empty name. 
//
// Set 'single' parameter to download only one favicon (see SelectSingleFavicon for the selection rule)
//
// Returns the slice of successfully downloded Favicons
func Download(website string, toDir string, single bool) (favicons []TFavicon, err error) {

    toDir = strings.ToLower(strings.Trim(toDir, " "))
    if len(toDir) == 0 {
        return favicons, fmt.Errorf("destination directory name should not be empty")
    }

    // create the dest dir
	if filepath.Base(toDir)[0] != '.' && !strings.HasPrefix(toDir, ".") {
		toDir = strings.TrimRight(toDir, "/") 
	}
    os.MkdirAll(toDir, 0755)

    // get the icons
    webfavicons, err := ReadAll(website)
    if err != nil {
        return favicons, err
    }

    if single {
        pone := SelectSingle(webfavicons)
        if pone != nil {
            webfavicons[0] = *pone
            webfavicons = webfavicons[:1]
        } else {
            // unable to select a single one, make sure favicons is empty
            webfavicons = webfavicons[:0]
        }
    }

    // save on disk each favicons
    nb := 0
    var outFile *os.File
    for _, favicon := range(webfavicons) {
        outFile, err = os.Create(filepath.Join(toDir, favicon.DiskFileName))
        if err != nil {
            fmt.Println(err)
            continue 
        }
        _, err = outFile.Write(favicon.Image)
        if err != nil {
            fmt.Println(err)
            outFile.Close()
            continue 
        }
        nb++
        outFile.Close()
        favicons = append(favicons, favicon)
    }
    return favicons, err
}
