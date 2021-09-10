package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"pvri.com/glctl/pkg/client"
)

func main() {
	// Create new parser object
	parser := argparse.NewParser("glctl", "GitLab control client")

	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Verbose mode"})

	// group command
	groupCmd := parser.NewCommand("group", "Group related commands")

	groupcloneCmd := groupCmd.NewCommand("clone", "Clone a group")

	group_src := groupcloneCmd.String("s", "source", &argparse.Options{Required: true, Help: "The source group"})
	group_dst := groupcloneCmd.String("d", "dest", &argparse.Options{Required: true, Help: "The dest parent group full path"})
	group_name := groupcloneCmd.String("n", "name", &argparse.Options{Required: true, Help: "The dest group name"})
	group_path := groupcloneCmd.String("p", "path", &argparse.Options{Help: "The dest group path"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		return
	}

	gl, err := client.NewClient("5Zsfo9xmqD5Mn1PfMrGq", "https://st-gitlab-dgt.eni.com", *verbose)

	if groupcloneCmd.Happened() {
		if group_src == nil ||
			group_dst == nil ||
			group_name == nil ||
			group_path == nil {
			fmt.Print(parser.Usage(err))
			return
		}
		err = gl.GroupClone(*group_src, *group_dst, *group_name, *group_path)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
