package clang_format

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const filename = ".clang-format"
const dot = "."
const TargetDirectory = "."
const UnitDirectory = "unit/"

var filePatterns = []string{
	"src/*.c",
	"src/*.h",
	"src/**/*.c",
	"src/**/*.h",
}

var options = map[string][]string{
	"Language":                                     {"Cpp"},           // 1
	"AlignConsecutiveAssignments.Enabled":          {"true", "false"}, // 2 - total 2
	"AlignConsecutiveAssignments.AcrossEmptyLines": {"true", "false"}, // 2 - total 4
	"AlignConsecutiveAssignments.AcrossComments":   {"true", "false"}, // 2 - total 8
	"AlignConsecutiveAssignments.AlignCompound":    {"true", "false"}, // 2 - total 16
	"AlignConsecutiveAssignments.PadOperators":     {"true", "false"}, // 2 - total 32
}

type ClangFormat map[string]string

func (c ClangFormat) String() string {
	groups := make(map[string]map[string]string)
	lines := make(map[string]string)

	// Sort everything into lines and groups
	for k, v := range c {
		// does this key have a .?
		if strings.Contains(k, dot) {
			parts := strings.Split(k, dot)
			if groups[parts[0]] == nil {
				groups[parts[0]] = make(map[string]string)
			}

			groups[parts[0]][parts[1]] = v
		} else {
			lines[k] = v
		}
	}

	buf := bytes.Buffer{}

	// Create an alphabetical list of non-grouped options
	linesAlphabetical := make([]string, len(lines))
	i := 0
	for lineKey := range lines {
		linesAlphabetical[i] = lineKey
		i++
	}

	// Sort alphabetical for the non-grouped options
	slices.Sort(linesAlphabetical)

	// Add the options to the buffer for the non-grouped options
	for _, key := range linesAlphabetical {
		buf.WriteString(
			fmt.Sprintf("%s: %s\n\n", key, lines[key]),
		)
	}

	// On to groups!
	// Create an alphabetical list of group names
	groupNamesAlphabetical := make([]string, len(groups))
	i = 0
	for key := range groups {
		groupNamesAlphabetical[i] = key
		i++
	}

	// Sort group names alphabetical
	slices.Sort(groupNamesAlphabetical)

	// For each group
	for _, groupKey := range groupNamesAlphabetical {
		// ... write the group name first, and then
		buf.WriteString(fmt.Sprintf("%s:\n", groupKey))

		// Sort the member keys into alphabetical order
		memberAlphabetical := make([]string, len(groups[groupKey]))
		g := 0

		for k := range groups[groupKey] {
			memberAlphabetical[g] = k
			g++
		}

		// Sort the group member names alphabetical
		slices.Sort(memberAlphabetical)

		// Write the member options in alphabetical order into the
		// buffer
		for _, memberKey := range memberAlphabetical {
			buf.WriteString(fmt.Sprintf("  %s: %s\n",
				memberKey,
				groups[groupKey][memberKey],
			))
		}

		// Close the group with an added newline
		buf.WriteString("\n")
	}

	// return the entire buffer
	return buf.String()
}

func IdealClangFormatFile() (ClangFormat, int, error) {
	format := generateBasic(options)

	// let's create a slice of option names
	optionNames := make([]string, len(options))
	i := 0
	for opt := range options {
		optionNames[i] = opt
		i++
	}

	// sort list alphabetically
	slices.Sort(optionNames)

	linesChangedTotal := math.MaxInt32

	// for each option, let's check whether their individual values
	// would produce a diff that has a lower changed line count.
	for _, optionName := range optionNames {
		if len(options[optionName]) < 2 {
			continue
		}

		changes := make(map[string]int)

		for _, value := range options[optionName] {
			format[optionName] = value

			linesChanged, err := runOption(format)
			if err != nil {
				return nil, 0, errors.Wrap(err, "runOption")
			}

			changes[value] = linesChanged
		}

		minLinesChanged := math.MaxInt32
		winningValue := ""
		for value, linesChanged := range changes {
			if linesChanged < minLinesChanged {
				winningValue = value
				minLinesChanged = linesChanged
			}
		}

		if linesChangedTotal > minLinesChanged {
			linesChangedTotal = minLinesChanged
		}

		format[optionName] = winningValue
	}

	return format, linesChangedTotal, nil
}

func runOption(option ClangFormat) (int, error) {
	fmt.Println("Writing .clang-format file")

	err := os.WriteFile(
		path.Join(TargetDirectory, filename),
		[]byte(option.String()),
		0755,
	)
	if err != nil {
		return 0, errors.Wrap(err, "os.WriteFile %04d")
	}

	var stdErr strings.Builder
	var stdOut strings.Builder

	CFCtx, CFCxl := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer CFCxl()

	clangFormatCmd := exec.CommandContext(CFCtx,
		"clang-format",
		"-i",
		"--style=file", // will find a .clang-format file going up the tree.
		"--verbose",
		"--files=files.list",
	)

	clangFormatCmd.Stderr = &stdErr
	// clangFormatCmd does not need stdOut

	fmt.Println("Running clang-format command")
	err = clangFormatCmd.Run()
	if err != nil {
		return 0, errors.Wrapf(err, "clangFormatCmd.Run(): %s", stdErr.String())
	}

	// let's get the diff
	diffCtx, diffCxl := context.WithTimeout(context.Background(), 10*time.Second)
	defer diffCxl()

	// Let's reset both writers, even though stdOut was not used above, probably
	stdOut.Reset()
	stdErr.Reset()

	diffCmd := exec.CommandContext(diffCtx,
		"git",
		"--no-pager",
		"diff",
		"--numstat",
	)
	diffCmd.Dir = UnitDirectory
	diffCmd.Stdout = &stdOut
	diffCmd.Stderr = &stdErr

	fmt.Println("Getting diff")
	err = diffCmd.Run()
	if err != nil {
		return 0, errors.Wrapf(err, "diff: %s", stdErr.String())
	}

	linesChanged, err := parseNumStat(stdOut.String())
	if err != nil {
		return 0, errors.Wrap(err, "parseNumStat")
	}

	fmt.Printf("Got diff, lines changed is %d\n", linesChanged)

	resetCtx, resetCxl := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer resetCxl()

	fmt.Printf("Resetting repository\n\n")
	resetCmd := exec.CommandContext(resetCtx,
		"git",
		"--no-pager",
		"reset",
		"--hard",
	)
	resetCmd.Dir = UnitDirectory
	err = resetCmd.Run()
	if err != nil {
		return 0, errors.Wrap(err, "reset")
	}

	return linesChanged, nil
}

func parseNumStat(output string) (int, error) {
	totalLinesChanged := 0

	trimmed := strings.TrimSpace(output)
	if len(trimmed) == 0 {
		return math.MaxInt32, nil
	}

	// Split the input by lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return math.MaxInt32, nil
	}

	for _, line := range lines {
		// Split each line into columns
		columns := strings.Fields(line)

		// Convert the first and second columns to integers and sum them
		linesAdded, err := strconv.Atoi(columns[0])
		if err != nil {
			return 0, errors.Wrapf(err, "failed to convert number %s on"+
				" col1 to int in line %s", columns[0], line)
		}

		linesRemoved, err := strconv.Atoi(columns[1])
		if err != nil {
			return 0, errors.Wrapf(err, "failed to convert number %s on"+
				" col2 to int in line %s", columns[1], line)
		}

		// get the larger number, the idea being that 6 lines added, 5 lines
		// removed means 5 lines changed + 1 line added = 6 total lines changed.
		if linesAdded > linesRemoved {
			totalLinesChanged += linesAdded
		} else {
			totalLinesChanged += linesRemoved
		}
	}

	return totalLinesChanged, nil
}
