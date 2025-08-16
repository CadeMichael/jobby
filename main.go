package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const reset = "\033[0m"
const bold = "\033[1m"
const italic = "\033[3m"
const cRed = "\033[31m"
const cGreen = "\033[32m"
const cBlue = "\033[34m"

type hist struct {
	Histogram map[string]int
}

type api struct {
	AppId string `json:"app_id"`
	ApiId string `json:"api_id"`
}

func buildUrl(loc string, job string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(dir + "/.config/jobby/jobby.json")
	if err != nil {
		return "", err
	}

	var conf api
	if err := json.Unmarshal(content, &conf); err != nil {
		return "", err
	}

	url := "https://api.adzuna.com/v1/api/jobs/" +
		loc +
		"/histogram?app_id=" +
		conf.AppId +
		"&app_key=" +
		conf.ApiId +
		"&what=" +
		job

	return url, nil
}

func buildApiCall(url string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var h hist
	if err := json.Unmarshal(body, &h); err != nil {
		return 0, err
	}

	sals := make([]int, 0, len(h.Histogram))
	for k := range h.Histogram {
		ki, err := strconv.Atoi(k)
		if err != nil {
			return 0, err
		}
		sals = append(sals, ki)
	}

	sort.Ints(sals)
	slices.Reverse(sals)
	total := 0

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "| %ssalary%s\t| %sjobs%s\t|\n", bold+cGreen, reset, bold+cRed, reset)

	for _, sal := range sals {
		num_jobs := h.Histogram[strconv.Itoa(sal)]
		fmt.Fprintf(w, "| %s%d%s\t| %s%d%s\t|\n", italic+cGreen, sal, reset, italic+cRed, num_jobs, reset)
		total += num_jobs
	}

	w.Flush()

	return total, nil
}

func main() {
	fmt.Printf("%s[ Jobby ]%s\n", bold, reset)

	fmt.Print("country code -> ")
	var loc string
	fmt.Scanf("%s", &loc)

	fmt.Print("job keyword -> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}

	job := scanner.Text()
	jobSlice := strings.Split(job, " ")
	jobs := strings.Join(jobSlice, "%20")

	fmt.Printf("> searching for %s[%s]%s jobs in %s%s%s...\n", italic, strings.Join(jobSlice, ","), reset, italic, loc, reset)

	url, err := buildUrl(loc, jobs)
	if err != nil {
		fmt.Println(err)
		return
	}

	num_jobs, err := buildApiCall(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("total jobs -> %s%d%s\n", cBlue, num_jobs, reset)
}
