package web

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/gobuffalo/packr/v2"
	"github.com/juju/loggo"
	"github.com/patrickmn/go-cache"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

type TemplatePages struct {
	PrevURI string
	NextURI string

	Pages   []*TemplatePage
}
type TemplatePage struct {
	PageNum string
	PageURI string
	Active  bool
}

var logger *loggo.Logger
var templates *packr.Box
var globalSessions *pgstore.PGStore

var templateCache *cache.Cache

func Close() {
	globalSessions.Close()
}

func compileTemplates(filenames ...string) (*template.Template, error) {
	start := time.Now()
	var tmpl *template.Template

	filenamesStr := fmt.Sprintf("%s", filenames)

	// This gets tedious if the value is used several times in the same function.
	// You might do either of the following instead:
	if x, found := templateCache.Get(filenamesStr); found {
		tmpl := x.(*template.Template)

		elapsed := time.Since(start)
		logger.Tracef("compileTemplates(%s) [%s][HIT]", filenames, elapsed)
		return tmpl, nil
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)

	for _, filename := range filenames {
		if tmpl == nil {
			tmpl = template.New(filename)
		} else {
			tmpl = tmpl.New(filename)
		}

		b, err := templates.FindString(filename)
		if err != nil {
			return nil, err
		}

		mb, err := m.String("text/html", b)
		if err != nil {
			return nil, err
		}
		tmpl.Parse(string(mb))

		tmpl.Parse(string(mb))
	}

	templateCache.Set(filenamesStr, tmpl, cache.DefaultExpiration)

	elapsed := time.Since(start)
	logger.Tracef("compileTemplates(%s) [%s][MISS]", filenames, elapsed)
	return tmpl, nil
}

func Init(db string) {
	newLogger := loggo.GetLogger("web.web")
	logger = &newLogger

	gs, err := pgstore.NewPGStore(db, []byte("secret-key"))
	if err != nil {
		logger.Errorf(err.Error())
	}
	globalSessions = gs

	// Load Templates
	templates = packr.New("templates", "./templates")

	// init cache
	templateCache = cache.New(5*time.Minute, 10*time.Minute)
}

func makePagination(path string, curPage uint, maxPage uint, displayPages uint) (pages *TemplatePages) {
	newPages := &TemplatePages{}
	halfPages := displayPages/2

	if curPage > 1 {
		prevPage := curPage - 1
		newPages.PrevURI = path + "?page=" + strconv.Itoa(int(prevPage))
	}

	if maxPage <= displayPages {
		for i := uint(1); i <= maxPage; i++ {
			active := false
			if i == curPage {active = true}

			pageStr := strconv.Itoa(int(i))
			pageUri := path + "?page=" + pageStr

			newPages.Pages = append(newPages.Pages, &TemplatePage{pageStr, pageUri, active})
		}
	} else {
		var startingPage uint
		if curPage <= halfPages {
			startingPage = 1
		} else if curPage > maxPage - halfPages {
			startingPage = maxPage - displayPages + 1
		} else {
			startingPage = curPage - halfPages
		}
		for i := uint(0); i < displayPages; i++ {
			newPage := startingPage + i
			active := false
			if newPage == curPage {active = true}

			pageStr := strconv.Itoa(int(newPage))
			pageUri := path + "?page=" + pageStr

			newPages.Pages = append(newPages.Pages, &TemplatePage{pageStr, pageUri, active})
		}
	}

	if curPage < maxPage {
		nextPage := curPage + 1
		newPages.NextURI = path + "?page=" + strconv.Itoa(int(nextPage))
	}

	return newPages
}