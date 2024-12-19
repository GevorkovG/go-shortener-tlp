package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
)

type Req struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type Resp struct {
	ID    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

func (a *App) APIshortBatch(w http.ResponseWriter, r *http.Request) {

	var (
		originals []Req
		shorts    []Resp
		links     []*objects.Link
		userID    string
	)

	token := r.Context().Value(cookies.ContextUserKey).(string)

	userID, err := usertoken.GetUserID(token)
	if err != nil {
		userID = ""
	}

	err = json.NewDecoder(r.Body).Decode(&originals)
	if err != nil {
		log.Println("didn't decode body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, val := range originals {

		key := generateID()
		resp := Resp{
			ID:    val.ID,
			Short: fmt.Sprintf(a.cfg.ResultURL+"/%s", key),
		}
		link := &objects.Link{
			Short:    key,
			Original: val.URL,
			UserID:   userID,
		}

		shorts = append(shorts, resp)
		links = append(links, link)

	}

	if err := a.Storage.InsertLinks(links); err != nil {
		log.Println("Didn't insert to table")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(shorts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(response)
	if err != nil {
		return
	}

}
