package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	khaltiPath     = os.Getenv("KHALTI_ROOT")
	credentialPath = os.Getenv("GS_CRED_PATH")
)

type FeatureTrack struct {
	Number     string
	HashValue  string
	Title      string
	Author     string
	Deployable bool
	RC         bool
	Production bool
}

func main() {

	if len(khaltiPath) == 0 {
		printError("Export KHALTI_ROOT: khalti root path")
		os.Exit(-1)
	}

	if len(credentialPath) == 0 {
		printError("Export `GS_CRED_PATH`: google sheet credential json file path")
		os.Exit(-1)
	}

	fmt.Println(text.Bold.Sprint("⛅ Reading Sheets\n"))

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialPath))
	if err != nil {
		printError("Unable to connect to google sheets services")
		os.Exit(-1)
	}

	toPickList := readSheet(srv)

	if len(toPickList) == 0 {
		printError("Nothing to pick")
		return
	}

	prettyPrint(toPickList)
	startCherryPick(toPickList)
}

func readSheet(srv *sheets.Service) []FeatureTrack {
	rng, err := srv.Spreadsheets.Values.Get("1daRsScN4C049lCdl1Z94wFqddU5T5JfLRfaIo1FX8Sg", "Tracking!A1:J").Do()
	if err != nil {
		printError("Unable to read the sheet")
		os.Exit(-1)
	}

	readyToPick := []FeatureTrack{}

	if len(rng.Values) == 0 {
		return readyToPick
	}

	for i, row := range rng.Values {
		if i == 0 {
			continue
		}

		deployable := row[7] == "TRUE"
		rc := row[8] == "TRUE"
		production := row[9] == "TRUE"

		if !deployable || rc || production {
			continue
		}

		readyToPick = append(readyToPick, FeatureTrack{
			Number:     row[0].(string),
			Title:      row[1].(string),
			Author:     row[3].(string),
			HashValue:  row[5].(string),
			Deployable: deployable,
			RC:         rc,
			Production: production,
		})
	}

	return readyToPick
}

func prettyPrint(prs []FeatureTrack) {
	wr := table.NewWriter()
	bold := text.Bold.Sprint

	wr.AppendHeader(table.Row{bold("#"), bold("Number"), bold("Title"), bold("Author"), bold("Hash")})

	for i, row := range prs {

		length := len(row.Title)
		if length > 50 {
			length = 50
		}

		wr.AppendRow(table.Row{
			i + 1,
			row.Number,
			row.Title[:length],
			row.Author,
			row.HashValue[:10],
		})
	}

	fmt.Println(wr.Render())
}

func startCherryPick(prs []FeatureTrack) {

	minionPath := path.Join(khaltiPath, "minion_flutter")

	if err := os.Chdir(minionPath); err != nil {
		printError(fmt.Sprintf("Unable to change dir to %s, %v", minionPath, err))
		os.Exit(-1)
	}

	fmt.Println(text.Bold.Sprintf("\n📂 %s", minionPath))

	fmt.Println(text.Bold.Sprint("\n🍒 Picking\n"))

	for i, pr := range prs {

		lenOfTitleToShow := len(pr.Title)
		if lenOfTitleToShow > 100 {
			lenOfTitleToShow = 100
		}

		space := 100 - lenOfTitleToShow + 5

		formattedText := text.FgYellow.Sprintf("[%d] %s %s", i+1, pr.Number, pr.Title[:lenOfTitleToShow])
		fmt.Print(formattedText)

		picked := isAlreadyPicked(pr.Number)

		if picked {
			padding := strings.Repeat(" ", space)
			printSuccess(fmt.Sprintf("%s%s", padding, "🎉 Already Picked"))
		} else {
			pickByHash(pr.HashValue)
		}
	}

	fmt.Println(text.Bold.Sprint("\n🎉 🍒 Picking Completed "))
}

func isAlreadyPicked(number string) bool {
	cmd := exec.Command("git", "log", "--oneline", "--grep", fmt.Sprintf("(#%s)", number))
	out, err := cmd.Output()
	if err != nil {
		printError(fmt.Sprintf("Unable to run git command %v", err))
		os.Exit(-1)
	}

	return len(out) != 0
}

func pickByHash(hashValue string) {
	cmd := exec.Command("git", "cherry-pick", hashValue)

	if output, err := cmd.Output(); err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			printError("\n\t 😭 Conflict, need manual input")
			os.Exit(-1)
			return
		}

		if strings.Contains(string(output), "nothing to commit, working tree clean") {
			printSuccess(fmt.Sprintf("\n\tAlready picked skipping cherry-pick for %s", hashValue))
			return
		}

		printError(string(output))

	}

	printSuccess("\t🎉 Picked")

}

func printError(message string) {
	fmt.Println(text.FgRed.Sprint(message))
}

func printSuccess(message string) {
	fmt.Println(text.FgGreen.Sprint(message))
}
