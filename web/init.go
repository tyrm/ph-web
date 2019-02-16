package web

import (
	"html/template"
	"regexp"
	"strconv"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/gobuffalo/packr/v2"
	"github.com/juju/loggo"
)

type TemplateBreadcrumb struct {
	Text   string
	URL    string
	Active bool
}

type TemplateNavbar struct {
	Nodes    []*TempalteNavbarNode
	Username string
}

type TempalteNavbarNode struct {
	Text     string
	URL      string
	MatchStr string
	FAIcon   string

	Active   bool
	Disabled bool

	Children []*TempalteNavbarNode
}

type TemplatePages struct {
	PrevURI string
	NextURI string

	Pages []*TemplatePage
}
type TemplatePage struct {
	PageNum string
	PageURI string
	Active  bool
}

type TemplateVarLayout struct {
	NavBar     *TemplateNavbar
	AlertSuccess string
	AlertError string
	AlertWarn  string
	Username  string
}

var logger *loggo.Logger
var templates *packr.Box
var globalSessions *pgstore.PGStore

func Close() {
	globalSessions.Close()
}

func compileTemplates(filenames ...string) (*template.Template, error) {
	start := time.Now()
	var tmpl *template.Template

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

		tmpl.Parse(b)
	}

	elapsed := time.Since(start)
	logger.Tracef("compileTemplates(%s) [%s][MISS]", filenames, elapsed)
	return tmpl, nil
}

func Init(db string, box *packr.Box) {
	newLogger := loggo.GetLogger("web")
	logger = &newLogger

	gs, err := pgstore.NewPGStore(db, []byte("secret-key"))
	if err != nil {
		logger.Errorf(err.Error())
	}
	globalSessions = gs

	// Load Templates
	templates = box
}

func makeNavbar(path string) (navbar *TemplateNavbar) {
	newNavbar := &TemplateNavbar{
		Nodes: []*TempalteNavbarNode{
			{
				Text:     "Home",
				MatchStr: "^/web/$",
				URL:      "/web/",
			},
			{
				Text: "Admin",
				URL:  "#",
				Children: []*TempalteNavbarNode{
					{
						Text:     "Oauth Clients",
						MatchStr: "^/web/oauth-clients/.*$",
						FAIcon:   "desktop",
						URL:      "/web/oauth-clients/",
					},
					{
						Text:     "Registry",
						MatchStr: "^/web/registry/.*$",
						FAIcon:   "book",
						URL:      "/web/registry/",
					},
					{
						Text:     "Users",
						MatchStr: "^/web/users/.*$",
						FAIcon:   "user",
						URL:      "/web/users/",
					},
					{
						Text:     "Something else here",
						FAIcon:   "paw",
						URL:      "#",
						Disabled: true,
					},
				},
			},
		},
	}

	for i := 0; i < len(newNavbar.Nodes); i++ {
		if newNavbar.Nodes[i].MatchStr != "" {
			match, err := regexp.MatchString(newNavbar.Nodes[i].MatchStr, path)
			if err != nil {
				logger.Errorf("makeNavbar:Error matching regex: %v", err)
			}
			if match {
				newNavbar.Nodes[i].Active = true
			}

		}

		if newNavbar.Nodes[i].Children != nil {
			for j := 0; j < len(newNavbar.Nodes[i].Children); j++ {

				if newNavbar.Nodes[i].Children[j].MatchStr != "" {
					subMatch, err := regexp.MatchString(newNavbar.Nodes[i].Children[j].MatchStr, path)
					if err != nil {
						logger.Errorf("makeNavbar:Error matching regex: %v", err)
					}

					if subMatch {
						newNavbar.Nodes[i].Active = true
						newNavbar.Nodes[i].Children[j].Active = true
					}

				}

			}
		}
	}

	return newNavbar
}

func makePagination(path string, curPage uint, maxPage uint, displayPages uint) (pages *TemplatePages) {
	newPages := &TemplatePages{}
	halfPages := displayPages / 2

	if curPage > 1 {
		prevPage := curPage - 1
		newPages.PrevURI = path + "?page=" + strconv.Itoa(int(prevPage))
	}

	if maxPage <= displayPages {
		for i := uint(1); i <= maxPage; i++ {
			active := false
			if i == curPage {
				active = true
			}

			pageStr := strconv.Itoa(int(i))
			pageUri := path + "?page=" + pageStr

			newPages.Pages = append(newPages.Pages, &TemplatePage{pageStr, pageUri, active})
		}
	} else {
		var startingPage uint
		if curPage <= halfPages {
			startingPage = 1
		} else if curPage > maxPage-halfPages {
			startingPage = maxPage - displayPages + 1
		} else {
			startingPage = curPage - halfPages
		}
		for i := uint(0); i < displayPages; i++ {
			newPage := startingPage + i
			active := false
			if newPage == curPage {
				active = true
			}

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
