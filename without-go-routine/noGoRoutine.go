package notmain

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"os"
	"encoding/csv"

	"github.com/PuerkitoBio/goquery"
)

var baseUrl string = "https://sg.indeed.com/jobs?q=golang&limit=50"

type jobCard struct {
	jobLink string
	jobTitle string
	companyName string
	companyLocation string
	salary string
	summary string
}

func main() {
	var jobs []jobCard
	totalPages := getPages()
	for i:=0; i<totalPages; i++ {
		extractedJobs := getPage(i)
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
}

func writeJobs(jobs []jobCard) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "Name", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range(jobs) {
		jobSlice := []string{"https://sg.indeed.com" + job.jobLink, job.jobTitle, job.companyName, job.companyLocation, job.salary, job.summary}
		wErr := w.Write(jobSlice)
		checkErr(wErr)
	}
	fmt.Println("Done extracting", len(jobs), "jobs")
}

func getPages() int {
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

func getPage(index int) []jobCard {
	pageUrl := baseUrl + "&" + "start=" + strconv.Itoa(index*50)
	var jobs []jobCard
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	
	searchCards := doc.Find(".tapItem")
	searchCards.Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobs = append(jobs, job)
	})
	return jobs
}

func extractJob(card *goquery.Selection) jobCard {
	var salary = ""
	link, _ := card.Attr("href")
	jobTitle := card.Find(".jobTitle").Text()
	companyName := card.Find(".companyName").Text()
	companyLocation := card.Find(".companyLocation").Text()
	salary = card.Find(".salary-snippet").Text()
	summary := card.Find(".job-snippet>ul>li").Text()
	return jobCard{
		jobLink: link,
		jobTitle: jobTitle,
		companyName: companyName,
		companyLocation: companyLocation,
		salary: salary,
		summary: summary,
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
