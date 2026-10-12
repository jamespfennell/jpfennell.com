// Command node serves the status page of one of my servers (a.jamesfennell.dev etc.).
//
// The page template is rendered by Hugo from layouts/node-template, so the page
// has the same design as the rest of the website. See README.md.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for one node.
type Config struct {
	// Hostname of the node; e.g. a.jamesfennell.dev.
	Hostname string `yaml:"hostname"`
	// Public IP address of the node.
	IP string `yaml:"ip"`
	// Where the node is running; e.g. "DigitalOcean, New York".
	Location string `yaml:"location"`
	// Entries to show on the page.
	Entries []Entry `yaml:"entries"`
	// Paths to files that each contain a list of entries.
	//
	// Paths are relative to the directory containing this config file.
	// Entries loaded from these files are added to Entries, in order.
	Include []string `yaml:"include"`
}

// Entry is one thing running on, or related to, the node.
type Entry struct {
	// Section of the page the entry appears in: one of the keys of sections.
	Section string `yaml:"section"`
	// Name of the entry; usually a hostname.
	Name string `yaml:"name"`
	// URL the name links to.
	URL string `yaml:"url"`
	// Optional source repository, without the scheme; e.g. github.com/jamespfennell/hoard.
	Repo string `yaml:"repo"`
}

// RepoName returns the name of the entry's repository, without the owner.
func (e Entry) RepoName() string {
	return path.Base(e.Repo)
}

// sections are the sections of the page, in order.
var sections = []struct {
	key   string
	title string
	intro string
}{
	{"infrastructure", "Infrastructure", ""},
	{"services", "Services", ""},
	{"global", "Global infrastructure", "Not node specific."},
}

type section struct {
	Title string
	// Optional paragraph shown under the title.
	Intro   string
	Entries []Entry
	// Whether any entry has a repo, in which case the section has a repo column.
	HasRepo bool
}

type page struct {
	Hostname string
	IP       string
	Location string
	Sections []section
}

func main() {
	configPath := flag.String("config", "", "path to the node's config file")
	templatePath := flag.String("template", "template.html", "path to the page template rendered by Hugo")
	staticDir := flag.String("static", "static", "directory containing the site's stylesheets")
	addr := flag.String("addr", ":8000", "address to listen on")
	flag.Parse()

	tmpl, err := template.New(filepath.Base(*templatePath)).Delims("[[", "]]").ParseFiles(*templatePath)
	if err != nil {
		log.Fatalf("failed to parse template: %s", err)
	}
	// Fail on startup if the config is broken. After that the config is
	// reloaded on every request, so config changes don't need a restart.
	if _, err := loadPage(*configPath); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	static := http.FileServer(http.Dir(*staticDir))
	http.Handle("/normalize.css", static)
	http.Handle("/stylesheet.css", static)
	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		p, err := loadPage(*configPath)
		if err != nil {
			log.Printf("failed to load config: %s", err)
			http.Error(w, fmt.Sprintf("failed to load config: %s", err), http.StatusInternalServerError)
			return
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, p); err != nil {
			log.Printf("failed to render page: %s", err)
			http.Error(w, "failed to render page", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b.Bytes())
	})
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func loadPage(configPath string) (*page, error) {
	var c Config
	if err := readYAML(configPath, &c); err != nil {
		return nil, err
	}
	entries := c.Entries
	for _, include := range c.Include {
		path := filepath.Join(filepath.Dir(configPath), include)
		var included []Entry
		if err := readYAML(path, &included); err != nil {
			return nil, err
		}
		entries = append(entries, included...)
	}
	p := &page{Hostname: c.Hostname, IP: c.IP, Location: c.Location}
	keys := map[string]bool{}
	for _, s := range sections {
		keys[s.key] = true
		sec := section{Title: s.title, Intro: s.intro}
		for _, e := range entries {
			if e.Section == s.key {
				sec.Entries = append(sec.Entries, e)
				sec.HasRepo = sec.HasRepo || e.Repo != ""
			}
		}
		p.Sections = append(p.Sections, sec)
	}
	for _, e := range entries {
		if !keys[e.Section] {
			return nil, fmt.Errorf("entry %q has unknown section %q", e.Name, e.Section)
		}
	}
	return p, nil
}

func readYAML(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err := d.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
