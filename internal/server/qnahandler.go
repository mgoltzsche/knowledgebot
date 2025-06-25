package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mgoltzsche/knowledgebot/internal/qna"
)

func newQuestionAnswerHandler(ai *qna.QuestionAnswerWorkflow) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		question := req.URL.Query().Get("q")
		if question == "" {
			err := req.ParseForm()
			if err != nil {
				slog.Warn("parse form data: " + err.Error())
			}

			question = req.Form.Get("q")
			if question == "" {
				http.Error(w, "parameter q not specified", http.StatusBadRequest)
				return
			}
		}

		ch, err := ai.Answer(req.Context(), question)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setHeaders(w.Header())

		for chunk := range ch {
			if chunk.Err != nil {
				chunk.Err = exposedError(chunk.Err.Error())
			}

			data, err := json.Marshal(chunk)
			if err != nil {
				slog.Error("failed to marshal chunk: " + err.Error())
				continue
			}

			if chunk.Err != nil {
				_, _ = fmt.Fprintln(w, "event: error")
			}

			_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	})
}

type exposedError string

func (e exposedError) Error() string {
	return string(e)
}

func setHeaders(h http.Header) {
	h.Set("Content-Type", "text/event-stream")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("X-Accel-Buffering", "no") // tell reverse proxy not to buffer
	h.Set("Connection", "keep-alive")

	// Set headers to prevent the client from caching
	h.Set("Cache-Control", "no-cache")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}
