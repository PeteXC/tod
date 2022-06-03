package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func InitialModel() model {
	return model{
		// Our shopping list is a grocery list
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func findOrCreateTodFile(filePath string) (err error) {
	s := fmt.Sprintf
	p := fmt.Println
	if _, err := os.Stat(filePath); err == nil {
		p(s("Found %s", filePath))
		return nil

	} else if errors.Is(err, os.ErrNotExist) {
		p(s("%s doesn't exist, trying to create ...", filePath))
		_, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("Something went wrong trying to create")
		}
	} else {
		return fmt.Errorf("Something went wrong, could not find %s", filePath)
	}
	return
}

func recursiveCreateAbsolutePath(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func getOrSetEnv(name, value string) (envValue string, err error) {
	envValue = os.Getenv(name)
	if envValue == "" {
		err = os.Setenv(name, value)
		if err != nil {
			return "", fmt.Errorf(fmt.Sprintf("Failed to set environment variable %s with %v", name, err))
		}
		return value, nil
	}
	return value, nil
}

func main() {
	fmt.Println(os.Getenv("HOME"))
	// First find or create the $TOD_HOME environment variable
	todHome, err := getOrSetEnv("TOD_HOME", os.Getenv("HOME")+"/.tod")
	if err != nil {
		panic(err)
	}

	fmt.Println(todHome)

	err = recursiveCreateAbsolutePath(todHome)
	check(err)

	// Find or create the config file
	err = findOrCreateTodFile(todHome + "/config.yml")
	if err != nil {
		panic(err)
	}
	// Find or create the data file
	err = findOrCreateTodFile(todHome + "/data.json")
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(InitialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
