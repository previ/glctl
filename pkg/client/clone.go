package client

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/previ/go-gitlab"
)

func (glc *GitLabClient) GroupClone(source string, dest string, name string, path string) error {

	if glc.verbose {
		glc.pi.CreateBar(name)
	}

	_, _, err := glc.client.Groups.GroupExportRequest(source)
	if err != nil {
		log.Fatal(err)
		return err
	}

	gsource, _, err := glc.client.Groups.GetGroup(source)
	if err != nil {
		log.Fatal(err)
		return err
	}

	gdestparent, _, err := glc.client.Groups.GetGroup(dest)
	if err != nil {
		log.Fatal(err)
		return err
	}

	var content []byte
	duration, _ := time.ParseDuration("10s")
	for i := 0; i < 60; i++ {
		var resp *gitlab.Response
		content, resp, err = glc.client.Groups.GroupExportDownload(source)
		if err != nil {
			if resp.StatusCode != 404 && resp.StatusCode != 429 {
				return err
			}
		}
		if resp.StatusCode == 200 {
			break
		}
		if glc.verbose {
			glc.pi.IncrementBar(name)
		}
		time.Sleep(duration)
	}
	tmpfile, err := ioutil.TempFile("", "export.*.tar.gz")
	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Fatal(err)
		return err
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
		return err
	}
	defer os.Remove(tmpfile.Name()) // clean up

	opt := gitlab.GroupImportOptions{
		Name:     gitlab.String(name),
		Path:     gitlab.String(path),
		ParentID: gitlab.String(strconv.Itoa(gdestparent.ID)),
		File:     gitlab.String(tmpfile.Name()),
	}
	_, _, err = glc.client.Groups.GroupImport(&opt)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		if glc.verbose {
			glc.pi.IncrementBar(name)
		}
	}
	gdest, _, err := glc.client.Groups.GetGroup(dest + "/" + path)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if glc.verbose {
		glc.pi.CompleteBar(name)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go glc.GroupDeepCopy(gdest, gsource, &wg)

	wg.Wait()

	return err
}

func (glc *GitLabClient) GroupDeepCopy(dest *gitlab.Group, source *gitlab.Group, wg *sync.WaitGroup) error {
	// for each subgroup
	//   copy avatar
	//   copy project
	//   deepcopy subgroup
	defer wg.Done()

	err := glc.CloneProjects(dest, source, wg)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	subgroups, _, err := glc.client.Groups.ListSubgroups(source.ID, &gitlab.ListSubgroupsOptions{})
	for _, sg := range subgroups {
		dg, _, err := glc.client.Groups.GetGroup(dest.FullPath + "/" + sg.Path)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		wg.Add(1)
		go glc.GroupDeepCopy(dg, sg, wg)
	}

	return nil
}

func (glc *GitLabClient) CopyAvatar(dest *gitlab.Group, source *gitlab.Group) error {
	return nil
}

func (glc *GitLabClient) CloneProjects(dest *gitlab.Group, source *gitlab.Group, wg *sync.WaitGroup) error {
	// for each project
	//   export project
	//   download project export
	//   import project export
	//	 copy avatar

	projects, _, err := glc.client.Groups.ListGroupProjects(source.ID, &gitlab.ListGroupProjectsOptions{})
	for _, ps := range projects {
		wg.Add(1)
		go glc.CloneProject(dest, ps, wg)
	}
	return err
}

func (glc *GitLabClient) CloneProject(dest *gitlab.Group, ps *gitlab.Project, wg *sync.WaitGroup) error {
	defer wg.Done()

	if glc.verbose {
		glc.pi.CreateBar(ps.Name)
	}

	_, err := glc.client.ProjectImportExport.ScheduleExport(ps.ID, &gitlab.ScheduleExportOptions{})
	if err != nil {
		log.Fatalln(err)
		return err
	}

	estatus := &gitlab.ExportStatus{}
	estatus, _, err = glc.client.ProjectImportExport.ExportStatus(ps.ID)
	for estatus.ExportStatus != "finished" {
		if err != nil {
			log.Fatalln(err)
			return err
		}
		time.Sleep(1 * time.Second)
		estatus, _, err = glc.client.ProjectImportExport.ExportStatus(ps.ID)

		if glc.verbose {
			glc.pi.IncrementBar(ps.Name)
		}
	}

	export, _, err := glc.client.ProjectImportExport.ExportDownload(ps.ID)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	tmpfile, err := ioutil.TempFile("", "export.*.tar.gz")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer os.Remove(tmpfile.Name()) // clean up
	if _, err := tmpfile.Write(export); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	opt := gitlab.ImportFileOptions{
		Name:      &ps.Name,
		Namespace: gitlab.String(dest.FullPath),
		File:      gitlab.String(tmpfile.Name()),
		Path:      gitlab.String(ps.Path),
	}

	istatus, _, err := glc.client.ProjectImportExport.ImportFile(&opt)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	istatus, _, err = glc.client.ProjectImportExport.ImportStatus(istatus.ID)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	for istatus.ImportStatus != "finished" {
		time.Sleep(1 * time.Second)
		istatus, _, err = glc.client.ProjectImportExport.ImportStatus(istatus.ID)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		if glc.verbose {
			glc.pi.IncrementBar(ps.Name)
		}
	}
	if glc.verbose {
		glc.pi.CompleteBar(ps.Name)
	}

	return nil
}
