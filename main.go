package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-android-unit-test/testaddon"
	"github.com/bitrise-tools/go-steputils/stepconf"
	shellquote "github.com/kballard/go-shellquote"
)

const resultArtifactPathPattern = "*TEST*.xml"

type config struct {
	AdditionalParams string `env:"additional_params"`
	ProjectLocation  string `env:"project_location,dir"`
}

func failf(msg string, args ...interface{}) {
	log.Errorf(msg, args...)
	os.Exit(1)
}

// func getArtifact(started time.Time, cfg config, pattern string) (artifact string, err error) {
// 	for _, t := range []time.Time{started, time.Time{}} {

// 		matches, err := filepath.Glob(cfg.ProjectLocation + pattern)
// 		if err != nil {
// 			return
// 		}
// 		if len(matches) == 0 {
// 			if t == started {
// 				log.Warnf("No artifacts found with pattern: %s that has modification time after: %s", pattern, t)
// 				log.Warnf("Retrying without modtime check....")
// 				fmt.Println()
// 				continue
// 			}
// 			log.Warnf("No artifacts found with pattern: %s without modtime check", pattern)
// 			log.Warnf("If you have changed default report export path in your gradle files then you might need to change ReportPathPattern accordingly.")
// 		}
// 		if len(matches) > 0 {
// 			artifact := matches[0]
// 		}
// 	}
// 	return
// }

func main() {
	var cfg config
	started := time.Now()

	if err := stepconf.Parse(&cfg); err != nil {
		failf("Issue with input: %s", err)
	}
	stepconf.Print(cfg)

	tapDartCmd := command.New("brew tap dart-lang/dart").
		SetStdout(os.Stdout).
		SetStderr(os.Stderr)
	tapDartCmd.Run()

	installDartCmd := command.New("brew install dart").
		SetStdout(os.Stdout).
		SetStderr(os.Stderr)
	installDartCmd.Run()

	addToPathCmd := command.New("export PATH=\"$PATH\":\"$HOME/.pub-cache/bin\"").
		SetStdout(os.Stdout).
		SetStderr(os.Stderr)
	addToPathCmd.Run()

	additionalParams, err := shellquote.Split(cfg.AdditionalParams)

	if err != nil {
		failf("Failed to parse additional parameters, error: %s", err)
	}

	fmt.Println()
	log.Infof("Running test")

	testCmd := command.New("flutter", append([]string{"test", "--machine | tojunit > TEST-report.xml"}, additionalParams...)...).
		SetStdout(os.Stdout).
		SetStderr(os.Stderr).
		SetDir(cfg.ProjectLocation)

	// artifact, err := getArtifact(started, cfg, resultArtifactPathPattern)
	if err != nil {
		log.Warnf("Failed to find test result XMLs, error: %s", err)
	} else {
		if baseDir := os.Getenv("BITRISE_TEST_RESULT_DIR"); baseDir != "" {
			uniqueDir, err := getUniqueDir(cfg.ProjectLocation + "TEST-report.xml")
			if err != nil {
				log.Warnf("Failed to export test results for test addon: cannot get export directory for artifact (%s): %s", err)
				return
			}

			if err := testaddon.ExportArtifact(cfg.ProjectLocation+"TEST-report.xml", baseDir, uniqueDir); err != nil {
				log.Warnf("Failed to export test results for test addon: %s", err)
			}
			log.Printf("  Exporting test results to test addon successful [ %s ] ", baseDir)
		}
	}

	fmt.Println()
	log.Donef("$ %s", testCmd.PrintableCommandArgs())
	fmt.Println()

	if err := testCmd.Run(); err != nil {
		failf("Running command failed, error: %s", err)
	}
}
