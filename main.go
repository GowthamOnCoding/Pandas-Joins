package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var plugins = map[string]string{
	"vim-fugitive":       "https://github.com/tpope/vim-fugitive/archive/refs/heads/master.zip",
	"vim-airline":        "https://github.com/vim-airline/vim-airline/archive/refs/heads/master.zip",
	"vim-airline-themes": "https://github.com/vim-airline/vim-airline-themes/archive/refs/heads/master.zip",
	"vim-gitgutter":      "https://github.com/airblade/vim-gitgutter/archive/refs/heads/master.zip",
	"grep.vim":           "https://github.com/vim-scripts/grep.vim/archive/refs/heads/master.zip",
	"CSApprox":           "https://github.com/vim-scripts/CSApprox/archive/refs/heads/master.zip",
	"delimitMate":        "https://github.com/Raimondi/delimitMate/archive/refs/heads/master.zip",
	"tagbar":             "https://github.com/majutsushi/tagbar/archive/refs/heads/master.zip",
	"ale":                "https://github.com/dense-analysis/ale/archive/refs/heads/master.zip",
	"material.vim":       "https://github.com/kaicataldo/material.vim/archive/refs/heads/master.zip",
	"fzf":                "https://github.com/junegunn/fzf/archive/refs/heads/master.zip",
	"fzf.vim":            "https://github.com/junegunn/fzf.vim/archive/refs/heads/master.zip",
	"vim-misc":           "https://github.com/xolox/vim-misc/archive/refs/heads/master.zip",
	"vim-session":        "https://github.com/xolox/vim-session/archive/refs/heads/master.zip",
	"ultisnips":          "https://github.com/SirVer/ultisnips/archive/refs/heads/master.zip",
	"vim-snippets":       "https://github.com/honza/vim-snippets/archive/refs/heads/master.zip",
	"vim-go":             "https://github.com/fatih/vim-go/archive/refs/heads/master.zip",
}

func main() {
	baseDir := "C:\\Users\\meena\\.vim\\pack\\plugins\\start"

	for name, url := range plugins {
		fmt.Printf("Downloading %s...\n", name)
		zipPath, err := downloadFile(url, name+".zip")
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", name, err)
			continue
		}

		fmt.Printf("Extracting %s...\n", name)
		err = unzip(zipPath, baseDir, name)
		if err != nil {
			fmt.Printf("Error extracting %s: %v\n", name, err)
			continue
		}

		fmt.Printf("Test")
		os.Remove(zipPath)
	}

	fmt.Println("All plugins downloaded and extracted!")
}

func downloadFile(url string, dest string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return dest, nil
}

func unzip(src string, dest string, pluginName string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create the plugin directory if it doesn't exist
	pluginPath := filepath.Join(dest, pluginName)
	if err := os.MkdirAll(pluginPath, os.ModePerm); err != nil {
		return err
	}

	for _, f := range r.File {
		// Skip the base directory
		name := f.Name
		if index := strings.Index(name, "/"); index != -1 {
			name = name[index+1:]
		}

		if name == "" {
			continue
		}

		fpath := filepath.Join(pluginPath, name)

		if !strings.HasPrefix(fpath, filepath.Clean(pluginPath)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(fpath)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
