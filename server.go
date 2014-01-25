package main

import (
	"encoding/json"
	"fmt"
	"github.com/mantasmatelis/go-trie-url-route"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {

	route, params, pathMatched := s.Router.FindRouteFromURL(r.Method, r.URL)

	if route == nil && pathMatched {
		http.Error(w, "", 405)
		return
	}
	if route == nil {
		http.Error(w, "", 400)
		return
	}

	route.Func.(func(http.ResponseWriter, *http.Request, map[string]string))(w, r, params)

	w.Write([]byte{}) /* Force a 200 status if none sent */
}

type Server struct {
	Router route.Router
}

func main() {
	/* Set up routes */
	var router route.Router
	router.SetRoutes(
		/* Server */
		route.Route{"GET", "/ping", ping},
		route.Route{"GET", "/:AuthToken/:PartyId/registerParty", registerParty},
		route.Route{"GET", "/:AuthToken/:PartyId/update", update},

		/* Both */
		route.Route{"GET", "/:AuthToken/getParties", getParties},

		/* Client */
		route.Route{"GET", "/:AuthToken/:PartyId/getPlaylist", getPlaylist},
		route.Route{"GET", "/:AuthToken/:PartyId/:SongId/up", upVote},
		route.Route{"GET", "/:AuthToken/:PartyId/:SongId/down", downVote},
		route.Route{"GET", "/:AuthToken/:PartyId/:SongId/null", nullVote},
	)

	/* Define the server, run it */
	s := &Server{Router: router}
	hs := &http.Server{
		Addr:    ":3100",
		Handler: http.HandlerFunc(s.handleRequest),
	}
	parties = make(map[string]*Party)
	parties["potato"] = &Party{}

	hs.ListenAndServe()
}

type Party struct {
	Playlist string
	Events   []Event
}

type Event struct {
	Type   int /* +1, 0, -1 */
	UserId string
	SongId string
}

var parties map[string]*Party

func ping(w http.ResponseWriter, req *http.Request, params map[string]string) {

}

func registerParty(w http.ResponseWriter, req *http.Request, params map[string]string) {
	parties[params["PartyId"]] = &Party{}
}

func update(w http.ResponseWriter, req *http.Request, params map[string]string) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		http.Error(w, "Could not read update", 499)
		log.Print(err)
		return
	}

	parties[params["PartyId"]].Playlist = string(body)

	data, err := json.Marshal(parties[params["PartyId"]].Events)

	if err != nil {
		http.Error(w, "Could not marshal", 499)
		log.Print(err)
	}
	parties[params["PartyId"]].Events = make([]Event, 0)
	w.Write(data)
}

func getParties(w http.ResponseWriter, req *http.Request, params map[string]string) {
	res, err := http.DefaultClient.Get("https://graph.facebook.com/me/events?access_token=" + params["AuthToken"])
	if err != nil {
		http.Error(w, "Facebook blew up", 499)
		log.Print(err)
		return
	}
	resp, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		http.Error(w, "Not able to read", 499)
		log.Print(err)
		return
	}
	fmt.Printf("%s", resp)
	w.Write(resp)

}

func getPlaylist(w http.ResponseWriter, req *http.Request, params map[string]string) {
	party, ok := parties[params["PartyId"]]

	if !ok {
		http.Error(w, "No party", 404)
	}

	w.Write([]byte(party.Playlist))
}

func upVote(w http.ResponseWriter, req *http.Request, params map[string]string) {
	party, ok := parties[params["PartyId"]]

	if !ok {
		http.Error(w, "No party", 404)
	}
	party.Events = append(party.Events, Event{Type: 1, UserId: params["AuthToken"], SongId: params["SongId"]})
}

func downVote(w http.ResponseWriter, req *http.Request, params map[string]string) {
	party, ok := parties[params["PartyId"]]

	if !ok {
		http.Error(w, "No party", 404)
	}
	party.Events = append(party.Events, Event{Type: -1, UserId: params["AuthToken"], SongId: params["SongId"]})
}

func nullVote(w http.ResponseWriter, req *http.Request, params map[string]string) {
	party, ok := parties[params["PartyId"]]

	if !ok {
		http.Error(w, "No party", 404)
	}

	party.Events = append(party.Events, Event{Type: 0, UserId: params["AuthToken"], SongId: params["SongId"]})
}
