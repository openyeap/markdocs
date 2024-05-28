package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v2"
)

func NewSite(cfg string) Site {
	//load site settings
	return scanSite(cfg)

}
func (site *Site) Public() error {
	err := os.RemoveAll(site.DstDir)
	if err != nil {
		log.Println("remove destination dir", err.Error())
	}
	err = os.MkdirAll(site.DstDir, os.ModeDir)
	if err != nil {
		log.Println("create destination dir", err.Error())
	}
	site.loadLayout()

	data, err := json.MarshalIndent(site.getToc(site.Manifest), "", "    ")
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(site.DstDir, "toc.json"))
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(data)

	site.gen(site.Manifest)
	return nil
}
func (site *Site) getToc(pages []Page) []Item {

	less := func(i, j int) bool {

		return pages[j].Sort() > pages[i].Sort()
	}
	sort.Slice(pages, less)

	toc := []Item{}
	for _, page := range pages {
		item := Item{
			Path:  page.Path,
			Title: page.Title,
		}

		if len(page.Children) > 0 {
			item.Children = site.getToc(page.Children)
		}
		toc = append(toc, item)
	}
	return toc
}
func (site *Site) gen(pages []Page) {
	for _, page := range pages {

		dst := filepath.Join(site.DstDir, page.Path)

		switch page.Type {
		case "markdown":
			dir := filepath.Dir(dst)
			_, err := os.Stat(dir)
			if err != nil {
				err = os.MkdirAll(dir, os.ModeDir)
				if err != nil {
					continue
				}
			}
			destination, err := os.Create(dst)
			if err != nil {
				continue
			}
			defer destination.Close()

			data := make(map[string]interface{})
			data["Content"] = page.Content
			data["Title"] = page.Title
			data["Date"] = page.Date
			data["Meta"] = page.Metadata
			data["UrlPath"] = X{Site: *site, Page: page}
			data["Site"] = convertSite(site)
			err = site.Template.ExecuteTemplate(destination, page.Layout+".html", data)
			if value, has := page.Metadata["cp"]; has {
				v, y := value.(string)
				if y {
					copy(dst, filepath.Join(site.DstDir, v))
				}
			}
			if err != nil {
				log.Println(err.Error())
				destination.Write([]byte(err.Error()))
				continue
			}
		case "index":
			_, err := os.Stat(dst)
			if err != nil {
				err = os.MkdirAll(dst, os.ModeDir)
				if err != nil {
					continue
				}
			}
			destination, err := os.Create(filepath.Join(dst, "index.html"))
			if err != nil {
				continue
			}
			defer destination.Close()
			data := make(map[string]interface{})
			data["Content"] = page.Content
			data["Title"] = page.Title
			data["Date"] = page.Date
			data["Meta"] = page.Metadata
			data["UrlPath"] = X{Site: *site, Page: page}
			data["Site"] = convertSite(site)
			err = site.Template.ExecuteTemplate(destination, page.Layout+".html", data)
			if err != nil {
				log.Println(err.Error())
				destination.Write([]byte(err.Error()))
				continue
			}
			site.gen(page.Children)
			continue
		default:
			copy(filepath.Join(site.SrcDir, page.Path), dst)
		}

	}
}
func convertSite(site *Site) map[string]interface{} {
	data := make(map[string]interface{})
	data["Url"] = site.Url
	data["Title"] = site.Title
	data["Date"] = site.Date
	data["Manifest"] = site.Manifest
	data["Toc"] = site.Toc
	data["Meta"] = site.Metadata
	return data
}

func (site *Site) Server() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					site.Public()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add(filepath.Join(site.Root(), "layout"))
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(filepath.Join(site.Root(), "layout", "partials"))
	if err != nil {
		log.Fatal(err)
	}
	// err = watcher.Add(site.SrcDir)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	err = filepath.WalkDir(site.SrcDir, func(path string, d os.DirEntry, err error) error {

		if d.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
func scanSite(path string) Site {

	var site Site
	file, err := os.Open(path)
	if err != nil {
		return site
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return site
	}
	json.Unmarshal([]byte(data), &site.Setting)
	if site.Metadata == nil {
		site.Metadata = make(map[string]interface{})
	}
	site.DstDir, _ = filepath.Abs(filepath.Join(site.Root(), site.DstDir))
	site.SrcDir, _ = filepath.Abs(filepath.Join(site.Root(), site.SrcDir))
	site.Manifest = walkSrcDir(site, "")
	return site
}

func walkSrcDir(site Site, dir string) []Page {
	pages := []Page{}
	root := filepath.Join(site.SrcDir, dir)
	list, err := os.ReadDir(root)
	if err != nil {
		return pages
	}
	for _, d := range list {
		absPath := filepath.Join(root, d.Name())
		relPath := Rel(site, absPath)
		//排除用户指定目录
		if matched, err := filepath.Match(site.Exclude, relPath); matched || err != nil {
			continue
		}

		file, _ := d.Info()

		//如果是目录
		if d.IsDir() {
			page := Page{
				Title:    d.Name(),
				Date:     file.ModTime(),
				Path:     relPath,
				Layout:   "default",
				Type:     "index",
				Metadata: make(map[string]interface{}),
			}
			page.Children = walkSrcDir(site, relPath)
			pages = append(pages, page)
			continue
		}

		ext, i := FileExtesion(relPath)
		if ext == ".md" || ext == ".markdown" {
			// xxx ,err:=file.Open()
			f, err := os.Open(absPath)
			if err != nil {
				continue
			}
			defer f.Close()
			//load raw data
			data, err := io.ReadAll(f)
			content := string(data)
			if err != nil {
				continue
			}

			page := Page{
				Title:    d.Name()[0 : len(d.Name())-len(ext)],
				Date:     file.ModTime(),
				Path:     relPath[0:i] + ".html",
				Layout:   "default",
				Type:     "markdown",
				Metadata: make(map[string]interface{}),
			}

			if strings.HasPrefix(content, "---") {
				items := strings.Split(content, "---")
				err = yaml.Unmarshal([]byte(items[1]), &page.Metadata)
				if err != nil {
					log.Println(absPath, err.Error())
					continue
				}
				if value, has := page.Metadata["title"]; has {
					page.Title = value.(string)
				}
				if value, has := page.Metadata["layout"]; has {
					page.Layout = value.(string)
				}
				if value, has := page.Metadata["date"]; has {
					input := value.(string)
					date, err := parseTime(input)
					if err == nil {
						page.Date = date
					}
				}
				content = strings.Join(items[2:], "---")
				page.Content = string(mdToHTML([]byte(content)))
			} else {
				page.Content = string(mdToHTML(data))

			}
			pages = append(pages, page)

		} else {
			page := Page{
				Title:    d.Name()[0 : len(d.Name())-len(ext)],
				Date:     file.ModTime(),
				Path:     relPath,
				Layout:   "default",
				Type:     "file",
				Metadata: make(map[string]interface{}),
			}
			pages = append(pages, page)
		}
	}
	return pages
}
func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.Attributes | parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
func (site *Site) loadLayout() error {

	matchedLayout, err := filepath.Glob(filepath.Join(site.Root(), site.Layout, "*.html"))
	if err != nil {
		return err
	}
	matchedPartials, err := filepath.Glob(filepath.Join(site.Root(), site.Layout, "partials", "*.html"))
	if err != nil {
		return err
	}
	matched := append(matchedLayout, matchedPartials...)
	temp := template.Must(template.ParseFiles(matched...))
	site.Template = temp
	err = filepath.WalkDir(filepath.Join(site.Root(), "layout"), func(path string, d os.DirEntry, err error) error {

		if !d.IsDir() {
			if strings.HasSuffix(path, ".html") {
				return nil
			}
			rel_path, _ := filepath.Rel(filepath.Join(site.Root(), "layout"), path)
			copy(path, filepath.Join(site.DstDir, rel_path))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	_, err = os.Stat(dst)
	if err != nil {
		err = os.MkdirAll(filepath.Dir(dst), os.ModeDir)
		if err != nil {
			return 0, err
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
func parseTime(input string) (time.Time, error) {
	var date time.Time
	input = strings.Trim(input, " ")
	length := len(input)
	if length >= 19 {
		year, err := strconv.Atoi(input[0:4])
		if err != nil {
			return date, err
		}
		month, err := strconv.Atoi(input[5:7])
		if err != nil {
			return date, err
		}
		day, err := strconv.Atoi(input[8:10])
		if err != nil {
			return date, err
		}
		hour, err := strconv.Atoi(input[11:13])
		if err != nil {
			return date, err
		}
		mm, err := strconv.Atoi(input[14:16])
		if err != nil {
			return date, err
		}
		ss, err := strconv.Atoi(input[17:19])
		if err != nil {
			return date, err
		}
		return time.Date(year, time.Month(month), day, hour, mm, ss, 0, time.Local), nil

	} else if length >= 10 {
		year, err := strconv.Atoi(input[0:4])
		if err != nil {
			return date, err
		}
		month, err := strconv.Atoi(input[5:7])
		if err != nil {
			return date, err
		}
		day, err := strconv.Atoi(input[8:10])
		if err != nil {
			return date, err
		}
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
	}
	return date, errors.New("time parse failed")
}
