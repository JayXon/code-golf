package routes

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JRaspass/code-golf/cookie"
)

func ord(i int) string {
	switch i % 10 {
	case 1:
		if i%100 != 11 {
			return "st"
		}
	case 2:
		if i%100 != 12 {
			return "nd"
		}
	case 3:
		if i%100 != 13 {
			return "rd"
		}
	}
	return "th"
}

var tmpl = template.New("").Funcs(template.FuncMap{
	// NOTE Only handles 0 - 999,999
	"comma": func(i int) string {
		if i > 999 {
			return fmt.Sprintf("%d,%03d", i/1000, i%1000)
		}

		return strconv.Itoa(i)
	},
	"ord": ord,
})

func init() {
	// Tests run from the package directory, walk upwards until we find views.
	for {
		if _, err := os.Stat("views"); os.IsNotExist(err) {
			os.Chdir("..")
		} else {
			break
		}
	}

	if err := filepath.Walk("views", func(path string, _ os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			if b, err := ioutil.ReadFile(path); err != nil {
				return err
			} else {
				name := path[len("views/") : len(path)-len(".html")]
				tmpl = template.Must(tmpl.New(name).Parse(string(b)))
			}
		}

		return err
	}); err != nil {
		panic(err)
	}
}

// Render wraps common logic required for rendering a view to the user.
func Render(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	name, title string,
	data interface{},
) {
	header := w.Header()

	header.Set("Content-Language", "en")
	header.Set("Content-Type", "text/html; charset=utf-8")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("Content-Security-Policy",
		"base-uri 'none';"+
			"connect-src 'self';"+
			"default-src 'none';"+
			"form-action 'none';"+
			"font-src 'self';"+
			"frame-ancestors 'none';"+
			"img-src 'self' data: avatars.githubusercontent.com;"+
			"script-src 'self';"+
			"style-src 'self'",
	)

	args := struct {
		CommonCssPath, Login, LoginURL, Title string
		Data                                  interface{}
	}{
		CommonCssPath: commonCssPath,
		Data:          data,
		Title:         title,
	}

	if _, args.Login = cookie.Read(r); args.Login == "" {
		args.LoginURL = "//github.com/login/oauth/authorize?client_id=7f6709819023e9215205&scope=user:email&redirect_uri=https://code-golf.io/callback?redirect_uri%3D" + url.QueryEscape(url.QueryEscape(r.RequestURI))
	}

	w.WriteHeader(code)

	if err := tmpl.ExecuteTemplate(w, name, args); err != nil {
		panic(err)
	}
}
