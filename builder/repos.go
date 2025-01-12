//
// Copyright © 2016-2021 Solus Project <copyright@getsol.us>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package builder

import (
	"fmt"
	log "github.com/DataDrake/waterlog"
	"github.com/getsolus/libosdev/disk"
	"os"
	"path/filepath"
)

const (
	// BindRepoDir is where we make repos available from the host side
	BindRepoDir = "/hostRepos"
)

// addLocalRepo will try to add the repo and bind mount it into the target
func (p *Package) addLocalRepo(notif PidNotifier, o *Overlay, pkgManager *EopkgManager, repo *Repo) error {
	// Ensure the source exists too. Sorta helpful like that.
	if !PathExists(repo.URI) {
		return fmt.Errorf("Local repo does not exist")
	}

	mman := disk.GetMountManager()

	// Ensure the target mountpoint actually exists ...
	tgt := filepath.Join(o.MountPoint, BindRepoDir[1:], repo.Name)
	if !PathExists(tgt) {
		if err := os.MkdirAll(tgt, 00755); err != nil {
			return err
		}
	}

	// BindMount the directory into place
	if err := mman.BindMount(repo.URI, tgt); err != nil {
		return err
	}
	o.ExtraMounts = append(o.ExtraMounts, tgt)

	// Attempt to autoindex the repo
	if repo.AutoIndex {
		log.Debugf("Reindexing repository %s\n", repo.Name)

		command := fmt.Sprintf("cd %s/%s; %s", BindRepoDir, repo.Name, eopkgCommand("eopkg index --skip-signing ."))
		err := ChrootExec(notif, o.MountPoint, command)
		notif.SetActivePID(0)
		if err != nil {
			return err
		}
	} else {
		tgtIndex := filepath.Join(tgt, "eopkg-index.xml.xz")
		if !PathExists(tgtIndex) {
			log.Warnf("Repository index doesn't exist. Please index it to use it. %s\n", repo.Name)
		}
	}

	// Now add the local repo
	chrootLocal := filepath.Join(BindRepoDir, repo.Name, "eopkg-index.xml.xz")
	return pkgManager.AddRepo(repo.Name, chrootLocal)
}

func (p *Package) removeRepos(pkgManager *EopkgManager, repos []string) error {
	if len(repos) < 1 {
		return nil
	}
	for _, id := range repos {
		log.Debugf("Removing repository %s\n", id)
		if err := pkgManager.RemoveRepo(id); err != nil {
			return fmt.Errorf("Failed to remove repository %s, reason: %s\n", id, err)
		}
	}
	return nil
}

// addRepos will add the specified filtered set of repos to the rootfs
func (p *Package) addRepos(notif PidNotifier, o *Overlay, pkgManager *EopkgManager, repos []*Repo) error {
	if len(repos) < 1 {
		return nil
	}
	for _, repo := range repos {
		if repo.Local {
			log.Debugf("Adding local repo to system %s %s\n", repo.Name, repo.URI)

			if err := p.addLocalRepo(notif, o, pkgManager, repo); err != nil {
				return fmt.Errorf("Failed to add local repo to system %s, reason: %s\n", repo.Name, err)
			}
			continue
		}
		log.Debugf("Adding repo to system %s %s\n", repo.Name, repo.URI)
		if err := pkgManager.AddRepo(repo.Name, repo.URI); err != nil {
			return fmt.Errorf("Failed to add repo to system %s, reason: %s\n", repo.Name, err)
		}
	}
	return nil
}

// ConfigureRepos will attempt to configure the repos according to the configuration
// of the manager.
func (p *Package) ConfigureRepos(notif PidNotifier, o *Overlay, pkgManager *EopkgManager, profile *Profile) error {
	repos, err := pkgManager.GetRepos()
	if err != nil {
		return err
	}

	var removals []string

	// Find out which repos to remove
	if len(profile.RemoveRepos) == 1 && profile.RemoveRepos[0] == "*" {
		for _, r := range repos {
			removals = append(removals, r.ID)
		}
	} else {
		for _, r := range profile.RemoveRepos {
			removals = append(removals, r)
		}
	}

	if err := p.removeRepos(pkgManager, removals); err != nil {
		return err
	}

	var addRepos []*Repo

	if (len(profile.AddRepos) == 1 && profile.AddRepos[0] == "*") || len(profile.AddRepos) == 0 {
		for _, repo := range profile.Repos {
			addRepos = append(addRepos, repo)
		}
	} else {
		for _, id := range profile.AddRepos {
			addRepos = append(addRepos, profile.Repos[id])
		}
	}

	return p.addRepos(notif, o, pkgManager, addRepos)
}
