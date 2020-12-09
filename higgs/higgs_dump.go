package higgs

import (
	"bufio"
	"compress/bzip2"
	"context"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TODO implement a force flag and differential download to only get files we dont already have
func (h *Higgs) PopulateAllHistoricalSDE() error {

	const fuzzdumpindex = "https://www.fuzzwork.co.uk/dump/"

	const reg = "(sde-)[0-9]{8}(-TRANQUILITY\\/)"

	resp, err := soup.Get(fuzzdumpindex)
	if err != nil {
		return errors.Wrap(err, "failed to reach fuzzy steve")
	}
	idx := soup.HTMLParse(resp)

	pages := idx.Find("pre").FindAll("a")
	for _, a := range pages {
		match, _ := regexp.MatchString(reg, a.Text())
		if match {
			fmt.Printf("MATCH SDE %s, from %s\n", a.Text(), a.Text()[4:12])
			err = h.populateSDELink(fuzzdumpindex+a.Text(), a.Text()[4:12])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Higgs) populateSDELink(parent string, date string) (err error) {

	const sqlsreg = "(sql).(bz2)"

	resp, err := soup.Get(parent)
	if err != nil {
		return err
	}
	idx := soup.HTMLParse(resp)

	ct := context.TODO()
	g, ctx := errgroup.WithContext(ct)

	sqls := idx.Find("pre").FindAll("a")
	for i := range sqls {
		sql := sqls[i]
		match, _ := regexp.MatchString(sqlsreg, sql.Text())
		if match {
			g.Go(func() error {

				dir, err := ioutil.TempDir("", "higgs")
				defer os.RemoveAll(dir)
				if err != nil {
					return errors.Wrap(err, "failed to create temp directory")
				}

				tmpFn := filepath.Join(dir, strings.TrimSuffix(sql.Text(), ".bz2"))
				url := parent + sql.Text()
				err = h.downloadDecompressFile(ctx, tmpFn, url)
				if err != nil {
					return err
				}

				_, err = h.goop.MariaClient.Exec("CREATE DATABASE IF NOT EXISTS sde" + date)
				if err != nil {
					return errors.Wrap(err, "failed to create database")
				}

				// This is the part I dont like, I would prefer not to exec random files.
				cmd := exec.Command("mysql", "-u", h.sqlUser, "-p"+h.sqlPass, "-h", h.sqlHost, "sde"+date, "-e", "source " + tmpFn)
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				err = cmd.Run()
				if err != nil {
					fmt.Println(cmd.String())
					stder, _ := cmd.CombinedOutput()
					return errors.Wrap(err, "failed to restore mysql dump: " + string(stder))
				}

				return nil
			})
		}
	}

	return g.Wait()
}

// downloadDecompressFile will download a url to a local file, stripping the bz2 compression as it goes.
func (h *Higgs) downloadDecompressFile(ctx context.Context, filepath string, url string) error {

	fmt.Printf("Downloading %s\n", url)

	// Get the data
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Buffer in the body
	br := bufio.NewReader(resp.Body)
	// Decompression reader
	cr := bzip2.NewReader(br)

	// Write the body to file
	_, err = io.Copy(out, cr)
	return err

}
