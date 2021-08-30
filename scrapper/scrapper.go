package scrapper

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"os"
	"encoding/csv"

	"github.com/PuerkitoBio/goquery"
)

type jobCard struct {
	jobLink string
	jobTitle string
	companyName string
	companyLocation string
	salary string
	summary string
}

func Scrape(term string) {
	var baseUrl string = "https://sg.indeed.com/jobs?q=" + term + "&limit=50"
	var jobs []jobCard
	c := make(chan []jobCard)
	totalPages := getPages(baseUrl)

	for i:=0; i<totalPages; i++ {
		go getPage(i, baseUrl, c)
	}

	for i:=0;i<totalPages;i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPages(baseUrl string) int {
	pages := 0
	res, err := http.Get(baseUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	return pages
}

func getPage(index int, baseUrl string, mainC chan<- []jobCard) {
	var jobs []jobCard
	c := make(chan jobCard)

	pageUrl := baseUrl + "&start=" + strconv.Itoa(index*50)
	fmt.Println("Requesting", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	
	searchCards := doc.Find(".tapItem")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i:=0;i<searchCards.Length();i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- jobCard) {
	var salary = ""
	link, _ := card.Attr("href")
	jobTitle := card.Find(".jobTitle").Text()
	companyName := card.Find(".companyName").Text()
	companyLocation := card.Find(".companyLocation").Text()
	salary = card.Find(".salary-snippet").Text()
	summary := card.Find(".job-snippet>ul>li").Text()
	c<- jobCard{
		jobLink: link,
		jobTitle: jobTitle,
		companyName: companyName,
		companyLocation: companyLocation,
		salary: salary,
		summary: summary,
	}
}

func writeJobs(jobs []jobCard) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "Name", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://sg.indeed.com" + job.jobLink, job.jobTitle, job.companyName, job.companyLocation, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with status ", res.StatusCode)
	}
}
