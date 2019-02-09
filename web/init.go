package web

import (
	"strconv"

	"github.com/antonlindstrom/pgstore"
	"github.com/gobuffalo/packr/v2"
	"github.com/juju/loggo"
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

func Close() {
	globalSessions.Close()
}

func Init(db string) {
	newLogger := loggo.GetLogger("web.web")
	logger = &newLogger

	gs, err := pgstore.NewPGStore(db, []byte("secret-key"))
	if err != nil {
		logger.Errorf(err.Error())
	}
	globalSessions = gs


	templates = packr.New("templates", "./templates")
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