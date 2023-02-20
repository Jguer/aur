package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Jguer/aur"
	"github.com/Jguer/aur/rpc"
)

const (
	boldCode  = "\x1b[1m"
	resetCode = "\x1b[0m"
)

// UseColor determines if package will emit colors.
var UseColor = true //nolint

const (
	searchMode = "search"
	infoMode   = "info"
)

func getSearchBy(value string) aur.By {
	switch value {
	case "name":
		return aur.Name
	case "maintainer":
		return aur.Maintainer
	case "submitter":
		return aur.Submitter
	case "depends":
		return aur.Depends
	case "makedepends":
		return aur.MakeDepends
	case "optdepends":
		return aur.OptDepends
	case "checkdepends":
		return aur.CheckDepends
	case "provides":
		return aur.Provides
	case "conflicts":
		return aur.Conflicts
	case "replaces":
		return aur.Replaces
	case "keywords":
		return aur.Keywords
	case "groups":
		return aur.Groups
	case "comaintainers":
		return aur.CoMaintainers
	default:
		return aur.NameDesc
	}
}

func usage() {
	fmt.Println("Usage:", os.Args[0], "<opts>", "<command>", "<pkg(s)>")
	fmt.Println("Available commands:", "info, search")
	fmt.Println("Available opts:", "-by <Search for packages using a specified field>")

	flag.Usage()

	fmt.Println("Example:", "aur-cli -verbose -by name search python3.7")
}

func versionRequestEditor(ctx context.Context, req *http.Request) error {
	req.Header.Add("User-Agent", "aur-cli/v1")

	return nil
}

func main() {
	var (
		by          string
		aurURL      string
		verbose     bool
		jsonDisplay bool
	)

	flag.StringVar(&by, "by", "name-desc", "Search for packages using a specified field"+
		"\n (name/name-desc/maintainer/depends/makedepends/optdepends/checkdepends/"+
		"\n  provides/conflicts/replaces/keywords/groups/comaintainers)")
	flag.StringVar(&aurURL, "url", "https://aur.archlinux.org/", "AUR URL")
	flag.BoolVar(&verbose, "verbose", false, "display verbose information")
	flag.BoolVar(&jsonDisplay, "json", false, "display result as JSON")
	flag.Parse()

	if flag.NArg() < 2 {
		usage()

		os.Exit(1)
	}

	mode := flag.Arg(0)

	aurClient, err := rpc.NewClient(rpc.WithBaseURL(aurURL),
		rpc.WithRequestEditorFn(versionRequestEditor), rpc.WithLogFn(func(s ...any) {
			fmt.Fprintln(os.Stdout, append([]any{"[DEBUG]"}, s...)...)
		}))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	results, err := getResults(aurClient, by, mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}

	if jsonDisplay {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)

			os.Exit(1)
		}

		fmt.Println(string(output))
	} else {
		for i := range results {
			switch mode {
			case infoMode:
				printInfo(&results[i], os.Stdout, aurURL, verbose)
			case searchMode:
				printSearch(&results[i], os.Stdout)
			default:
				usage()
				os.Exit(1)
			}
		}
	}
}

func getResults(aurClient *rpc.Client, by, mode string) ([]aur.Pkg, error) {
	var (
		results []aur.Pkg
		err     error
	)

	switch mode {
	case searchMode:
		results, err = aurClient.Get(context.Background(), &aur.Query{
			Needles: flag.Args()[1:], By: getSearchBy(by), Contains: true,
		})
	case infoMode:
		results, err = aurClient.Get(context.Background(), &aur.Query{Needles: flag.Args()[1:]})
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		err = fmt.Errorf("rpc request failed: %w", err)
	}

	return results, err
}

func stylize(startCode, in string) string {
	if UseColor {
		return startCode + in + resetCode
	}

	return in
}

func Bold(in string) string {
	return stylize(boldCode, in)
}

func printSearch(a *aur.Pkg, w io.Writer) {
	fmt.Fprintf(w, "- %s %s (%d %.2f)\n\t%s\n",
		Bold(a.Name), a.Version, a.NumVotes, a.Popularity, a.Description)
}

// PrintInfo prints package info like pacman -Si.
func printInfo(pkg *aur.Pkg, writer io.Writer, aurURL string, verbose bool) {
	printInfoValue(writer, "Name", pkg.Name)
	printInfoValue(writer, "Version", pkg.Version)
	printInfoValue(writer, "Description", pkg.Description)

	if verbose {
		printInfoValue(writer, "Keywords", pkg.Keywords...)
		printInfoValue(writer, "URL", pkg.URL)
		printInfoValue(writer, "AUR URL", strings.TrimRight(aurURL, "/")+"/packages/"+pkg.Name)

		printInfoValue(writer, "Groups", pkg.Groups...)
		printInfoValue(writer, "Licenses", pkg.License...)
		printInfoValue(writer, "Provides", pkg.Provides...)
		printInfoValue(writer, "Depends On", pkg.Depends...)
		printInfoValue(writer, "Make Deps", pkg.MakeDepends...)
		printInfoValue(writer, "Check Deps", pkg.CheckDepends...)
		printInfoValue(writer, "Optional Deps", pkg.OptDepends...)
		printInfoValue(writer, "Conflicts With", pkg.Conflicts...)

		printInfoValue(writer, "Maintainer", pkg.Maintainer)
		printInfoValue(writer, "Submitter", pkg.Submitter)
		printInfoValue(writer, "Co-Maintainers", pkg.CoMaintainers...)
		printInfoValue(writer, "Votes", fmt.Sprintf("%d", pkg.NumVotes))
		printInfoValue(writer, "Popularity", fmt.Sprintf("%f", pkg.Popularity))
		printInfoValue(writer, "First Submitted", formatTimeQuery(pkg.FirstSubmitted))
		printInfoValue(writer, "Last Modified", formatTimeQuery(pkg.LastModified))

		if pkg.OutOfDate != 0 {
			printInfoValue(writer, "Out-of-date", formatTimeQuery(pkg.OutOfDate))
		} else {
			printInfoValue(writer, "Out-of-date", "No")
		}

		printInfoValue(writer, "ID", fmt.Sprintf("%d", pkg.ID))
		printInfoValue(writer, "Package Base ID", fmt.Sprintf("%d", pkg.PackageBaseID))
		printInfoValue(writer, "Package Base", pkg.PackageBase)
		printInfoValue(writer, "Snapshot URL", aurURL+pkg.URLPath)
	}

	fmt.Fprintln(writer)
}
