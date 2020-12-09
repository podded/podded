package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Server) home(c *gin.Context) {

	err, recents := s.goop.KillmailFetchRecent(c, 0)

	if err != nil {
		c.HTML(
			http.StatusInternalServerError,
			"page_error.html",
			gin.H{"error" : err.Error()},
			)
		return
	}

	c.HTML(
		http.StatusOK,
		"page_index.html",
		gin.H{
			"title": "Podded in Space!",
			"mails": recents,
		},
	)
}

func (s *Server) characterHome(c *gin.Context) {

	ids := c.Param("charid")
	id, err := strconv.Atoi(ids)

	if err != nil{
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	err, recents := s.goop.KillmailFetchCharacterRecent(c, id, 0)

	if err != nil {
		c.HTML(
			http.StatusInternalServerError,
			"page_error.html",
			gin.H{"error" : err.Error()},
		)
		return
	}

	c.HTML(
		http.StatusOK,
		"page_character.html",
		gin.H{
			"title": "Podded in Space!",
			"mails": recents,
			"charid": id,
		},
	)
}



func (s *Server) kill(c *gin.Context) {

	ids := c.Param("killid")
	id, err := strconv.Atoi(ids)

	if err != nil{
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	err, mail := s.goop.KillmailFetchIndividual(c, id)

	if err != nil {
		c.HTML(
			http.StatusInternalServerError,
			"page_error.html",
			gin.H{"error" : err.Error()},
		)
		return
	}

	c.HTML(
		http.StatusOK,
		"page_killmail.html",
		gin.H{
			"title": "Podded in Space!",
			"mail":  mail,
		},
	)
}


