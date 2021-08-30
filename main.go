package main

import (
	"strings"
	"os"

	"github.com/pyokang/job-scrapper/scrapper"
	"github.com/labstack/echo/v4"
  )

  const fileName string = "jobs.csv"

func main() {
	e := echo.New()

	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)

	e.Logger.Fatal(e.Start(":1323"))
}

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := c.FormValue("term")
	term = strings.ToLower(strings.TrimSpace(term))
	scrapper.Scrape(term)
	return c.Attachment(fileName, "job.csv")
}