package server

import (
	"github.com/gin-gonic/gin"
	"github.com/podded/podded/ectoplasma"
	"html/template"
)

type Server struct {
	goop *ectoplasma.PodGoo
	tmp *template.Template
}

func NewServer(goop *ectoplasma.PodGoo) *Server {
	return &Server{goop: goop}
}

func (s *Server)RunServer() error {

	// Create the default gin router, has logger and crash recovery middleware
	router := gin.Default()
	//Load the templates
	router.LoadHTMLGlob("server/templates/*")

	// Routes

	// INDEX
	router.GET("/", s.home)

	//KILLMAIL
	router.GET("/kill/:killid", s.kill)

	//CHARACTER
	router.GET("/character/:charid", s.characterHome)

	router.Run()

	return nil
}
