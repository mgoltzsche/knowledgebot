package server

import (
	"net/http"

	"github.com/mgoltzsche/knowledgebot/internal/qna"
)

type Routes struct {
	WebDir   string
	Workflow *qna.QuestionAnswerWorkflow
}

func (r *Routes) AddRoutes(mux *http.ServeMux) {
	mux.Handle("/", http.RedirectHandler("/ui/", http.StatusTemporaryRedirect))
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(r.WebDir))))
	mux.Handle("/api/qna", newQuestionAnswerHandler(r.Workflow))
}
