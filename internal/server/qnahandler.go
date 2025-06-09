package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mgoltzsche/knowledgebot/internal/qna"
)

func newQuestionAnswerHandler(ai *qna.QuestionAnswerWorkflow) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		question := req.URL.Query().Get("q")
		if question == "" {
			err := req.ParseForm()
			if err != nil {
				log.Println("WARNING: parse form data:", err)
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
			eventType := "chunk"
			if chunk.Err != nil {
				eventType = "error"
				chunk.Err = exposedError(chunk.Err.Error())
			}

			data, err := json.Marshal(chunk)
			if err != nil {
				log.Println("ERROR: failed to marshal chunk:", err)
				continue
			}

			fmt.Fprintf(w, "event: %s\n", eventType)
			fmt.Fprintf(w, "data: %s\n\n", string(data))

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

	// Set headers to prevent the client from caching
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}
