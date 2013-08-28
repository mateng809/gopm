// Copyright 2013 gopm authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package doc

import (
	"errors"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/Unknwon/com"
	"github.com/Unknwon/ctw/packer"
)

var (
	googlePattern = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

// getGoogleDoc downloads raw files from code.google.com.
func getGoogleDoc(client *http.Client, match map[string]string, installRepoPath string, nod *Node, cmdFlags map[string]bool) ([]string, error) {
	packer.SetupGoogleMatch(match)
	// Check version control.
	if err := packer.GetGoogleVCS(client, match); err != nil {
		return nil, err
	}

	var installPath string
	if nod.ImportPath == nod.DownloadURL {
		suf := "." + nod.Value
		if len(suf) == 1 {
			suf = ""
		}
		projectPath := expand("code.google.com/p/{repo}{dot}{subrepo}{dir}", match)
		installPath = installRepoPath + "/" + projectPath + suf
		nod.ImportPath = projectPath
	} else {
		installPath = installRepoPath + "/" + nod.ImportPath
	}

	// Remove old files.
	os.RemoveAll(installPath + "/")
	match["tag"] = nod.Value

	ext := ".zip"
	if match["vcs"] == "svn" {
		ext = ".tar.gz"
		com.ColorLog("[WARN] SVN detected, may take very long time.\n")
	}

	err := packer.PackToFile(match["importPath"], installPath+ext, match)
	if err != nil {
		return nil, err
	}

	var dirs []string
	if match["vcs"] != "svn" {
		dirs, err = com.Unzip(installPath+ext, path.Dir(installPath))
	} else {
		dirs, err = com.UnTarGz(installPath+ext, path.Dir(installPath))
	}

	if len(dirs) == 0 {
		return nil, errors.New("No file in repository")
	}

	if err != nil {
		return nil, err
	}
	os.Remove(installPath + ext)
	os.Rename(path.Dir(installPath)+"/"+dirs[0], installPath)

	// Check if need to check imports.
	if nod.IsGetDeps {
		imports := getImports(installPath+"/", match, cmdFlags, nod)
		return imports, err
	}

	return nil, err
}
