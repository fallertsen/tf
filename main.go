package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

var (
	ErrTooManyFiles = errors.New("Too many files in this sub-directory")
)

func PrintUsage() {
	fmt.Printf("Usage: tf <command> [args]\n\n")
	fmt.Printf("Available commands:\n")
	fmt.Printf("  status                     - Get the status of all the components\n")
	fmt.Printf("  output <component>         - Run the 'output' of the component\n")
	fmt.Printf("  plan <component>           - Run the 'plan' of the component\n")
	fmt.Printf("  apply <component> [-yes]   - Run the 'apply' of the component (-yes is the same as -auto-approve)\n")
	fmt.Printf("  destroy <component> [-yes] - Run the 'destroy' of the component (-yes is the same as -auto-approve)\n")
}

// InternalError is an error that is unexpected and should not happen.
func InternalError(msg string, err error) {
	fmt.Printf("Internal error: %s - %s", msg, err)
	os.Exit(2)
}

// Error is an error that can happen and we need to report it to the user.
func Error(msg string) {
	fmt.Printf("Error: %s\n", msg)
	os.Exit(1)
}

// FindAllComponents finds all the components in all the subfolders of the
// directory passed as argument. If we are going to scan too many files we are
// going to report an error, because it was probably not the intention of the
// user to run this command on that directory (for example the root directory).
func FindAllComponents(wd string) ([]string, error) {
	const MAX_FILES = 1_000

	components := []string{}

	numWalks := 0

	err := filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		numWalks += 1
		if numWalks > MAX_FILES {
			return ErrTooManyFiles
		}

		if err != nil {
			return err
		}

		if info.Name() != "main.tf" {
			return nil
		}

		// The component name should be the relative path between the
		// working directory and the main.tf.
		component := strings.TrimPrefix(path, wd)
		component = strings.TrimPrefix(component, "/")
		component = strings.TrimSuffix(component, "/main.tf")

		components = append(components, component)

		return nil
	})
	if err != nil {
		return []string{}, err
	}

	return components, nil
}

// GetStatus returns "destroyed" or "applied" depending on the status of the
// component.
func GetStatus(component string) string {
	tfstateFile := path.Join(component, "terraform.tfstate")

	if _, err := os.Stat(tfstateFile); os.IsNotExist(err) {
		return "destroyed"
	}

	tfstateBody, err := ioutil.ReadFile(tfstateFile)
	if err != nil {
		InternalError(fmt.Sprintf("GetStatus: Could not read the terraform.tfstate of component '%s'", component), err)
	}

	type tfState struct {
		Resources []struct {
			Type string `json:"type"`
		} `json:"resources"`
	}

	var s tfState
	err = json.Unmarshal(tfstateBody, &s)
	if err != nil {
		InternalError("GetStatus: Could not unmarshal terraform.tfstate", err)
	}

	if len(s.Resources) == 0 {
		return "destroyed"
	}

	return "applied"
}

// CmdStatus is run for the "status" command.
func CmdStatus() {
	wd, err := os.Getwd()
	if err != nil {
		InternalError("Could not find the current working directory", err)
	}

	components, err := FindAllComponents(wd)
	if err == ErrTooManyFiles {
		Error("We found more than 1000 files in the subdirectories, maybe you should try to run the command on a subdirectory with less files")
	}
	if err != nil {
		InternalError("FindAllComponents failed", err)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer writer.Flush()

	for _, component := range components {
		status := GetStatus(component)

		fmt.Fprintf(writer, "%s\t%s\n", component, status)
	}
}

// CmdStatus is run for the "output" command.
func CmdOutput() {
	component := os.Args[2]

	stat, err := os.Stat(component)
	if os.IsNotExist(err) {
		Error(fmt.Sprintf("Component '%s' not found", component))
	}
	if stat.IsDir() == false {
		Error(fmt.Sprintf("Component '%s' is not a folder", component))
	}

	cmd := exec.Command("terraform", "output")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = component
	cmd.Run()
}

// CmdStatus is run for the "plan" command.
func CmdPlan() {
	component := os.Args[2]

	stat, err := os.Stat(component)
	if os.IsNotExist(err) {
		Error(fmt.Sprintf("Component '%s' not found", component))
	}
	if stat.IsDir() == false {
		Error(fmt.Sprintf("Component '%s' is not a folder", component))
	}

	cmd := exec.Command("terraform", "plan")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = component
	cmd.Run()
}

// CmdStatus is run for the "apply" command.
func CmdApply() {
	component := os.Args[2]

	stat, err := os.Stat(component)
	if os.IsNotExist(err) {
		Error(fmt.Sprintf("Component '%s' not found", component))
	}
	if stat.IsDir() == false {
		Error(fmt.Sprintf("Component '%s' is not a folder", component))
	}

	cmd := exec.Command("terraform", "apply")
	if os.Args[3] == "-yes" {
		cmd = exec.Command("terraform", "apply", "-auto-approve")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = component
	cmd.Run()
}

// CmdStatus is run for the "destroy" command.
func CmdDestroy() {
	component := os.Args[2]

	stat, err := os.Stat(component)
	if os.IsNotExist(err) {
		Error(fmt.Sprintf("Component '%s' not found", component))
	}
	if stat.IsDir() == false {
		Error(fmt.Sprintf("Component '%s' is not a folder", component))
	}

	cmd := exec.Command("terraform", "destroy")
	if os.Args[3] == "-yes" {
		cmd = exec.Command("terraform", "destroy", "-auto-approve")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = component
	cmd.Run()
}

func main() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}

	if os.Args[1] == "status" {
		CmdStatus()
	} else if os.Args[1] == "output" {
		CmdOutput()
	} else if os.Args[1] == "plan" {
		CmdPlan()
	} else if os.Args[1] == "apply" {
		CmdApply()
	} else if os.Args[1] == "destroy" {
		CmdDestroy()
	} else {
		PrintUsage()
		os.Exit(1)
	}
}
