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

var errNoLinesChanged = errors.New("no lines changed")

var bools = []string{"true", "false"}

var options = map[string][]string{
	"AlignConsecutiveAssignments.Enabled":                bools, // 2 - total 2
	"AlignConsecutiveAssignments.AcrossEmptyLines":       bools, // 2 - total 4
	"AlignConsecutiveAssignments.AcrossComments":         bools, // 2 - total 8
	"AlignConsecutiveAssignments.AlignCompound":          bools, // 2 - total 16
	"AlignConsecutiveAssignments.PadOperators":           bools, // 2 - total 32
	"AlignAfterOpenBracket":                              {"Align", "DontAlign", "AlwaysBreak", "BlockIndent"},
	"AlignArrayOfStructures":                             {"None", "Left", "Right"},
	"AlignConsecutiveBitFields.Enabled":                  bools,
	"AlignConsecutiveBitFields.AcrossEmptyLines":         bools,
	"AlignConsecutiveBitFields.AcrossComments":           bools,
	"AlignConsecutiveDeclarations.Enabled":               bools,
	"AlignConsecutiveDeclarations.AcrossEmptyLines":      bools,
	"AlignConsecutiveDeclarations.AcrossComments":        bools,
	"AlignConsecutiveDeclarations.AlignFunctionPointers": bools,
	"AlignConsecutiveMacros.Enabled":                     bools,
	"AlignConsecutiveMacros.AcrossEmptyLines":            bools,
	"AlignConsecutiveMacros.AcrossComments":              bools,
	"AlignEscapedNewlines":                               {"Right", "Left", "LeftWithLastLine", "DontAlign"},
	"AlignOperands":                                      {"Align", "DontAlign", "AlignAfterOperator"},
	"AlignTrailingComments.Kind":                         {"Always", "Leave", "Never"},
	"AlignTrailingComments.OverEmptyLines":               {"0"},
	"AllowAllArgumentsOnNextLine":                        bools,
	"AllowAllParametersOfDeclarationOnNextLine":          bools,
	"AllowShortBlocksOnASingleLine":                      {"Always", "Empty", "Never"},
	"AllowShortCaseExpressionOnASingleLine":              bools,
	"AllowShortCaseLabelsOnASingleLine":                  bools,
	"AllowShortEnumsOnASingleLine":                       bools,
	"AllowShortFunctionsOnASingleLine":                   {"None", "InlineOnly", "Empty", "All", "Inline"},
	"AllowShortIfStatementsOnASingleLine":                {"Never", "WithoutElse", "OnlyFirstIf", "AllIfsAndElse"},
	"AllowShortLoopsOnASingleLine":                       bools,
	"AlwaysBreakBeforeMultilineStrings":                  bools,
	// "AttributeMacros" - this is not a multiple choice, but a list. This finder doesn't implement handling
	// lists
	"BinPackArguments":                    bools,
	"BinPackParameters":                   bools,
	"BitFieldColonSpacing":                {"None", "Both", "Before", "After"},
	"BraceWrapping.AfterCaseLabel":        bools,
	"BraceWrapping.AfterClass":            bools,
	"BraceWrapping.AfterControlStatement": {"Never", "MultiLine", "Always"},
	"BraceWrapping.AfterEnum":             bools,
	"BraceWrapping.AfterExternBlock":      bools,
	"BraceWrapping.AfterFunction":         bools,
	"BraceWrapping.AfterNamespace":        bools,
	"BraceWrapping.AfterObjCDeclaration":  bools,
	"BraceWrapping.AfterStruct":           bools,
	"BraceWrapping.AfterUnion":            bools,
	"BraceWrapping.BeforeCatch":           bools,
	"BraceWrapping.BeforeElse":            bools,
	"BraceWrapping.BeforeLambdaBody":      bools,
	"BraceWrapping.BeforeWhile":           bools,
	"BraceWrapping.IndentBraces":          bools,
	"BraceWrapping.SplitEmptyFunction":    bools,
	"BraceWrapping.SplitEmptyRecord":      bools,
	"BraceWrapping.SplitEmptyNamespace":   bools,
	"BreakAdjacentStringLiterals":         bools,
	"BreakAfterReturnType":                {"None", "Automatic", "ExceptShortType", "All", "TopLevel", "AllDefinitions", "TopLevelDefinitions"},
	"BreakBeforeBinaryOperators":          {"All", "None", "NonAssignment"},
	"BreakBeforeBraces":                   {"Custom"}, // This needs to be custom, so the BraceWrapping gets used
	"BreakBeforeTernaryOperators":         bools,
	"BreakConstructorInitializers":        {"BeforeColon", "BeforeComma", "AfterColon"},
	"BreakFunctionDefinitionParameters":   bools,
	"BreakStringLiterals":                 bools,
	"ColumnLimit":                         {"80"},
	"CommentPragmas":                      {"'^ IWYU pragma:'"},
	"ContinuationIndentWidth":             {"2"},
	"DerivePointerAlignment":              bools,
	"DisableFormat":                       {"false"},    // this needs to be false
	"IncludeBlocks":                       {"Preserve"}, // This needs to stay as is because it's brittle
	// all other include related rules are skipped
	"IndentAccessModifiers":                bools,
	"IndentCaseBlocks":                     bools,
	"IndentCaseLabels":                     bools,
	"IndentGotoLabels":                     bools,
	"IndentPPDirectives":                   {"None", "AfterHash", "BeforeHash"},
	"IndentWidth":                          {"4"},
	"IndentWrappedFunctionNames":           bools,
	"InsertNewlineAtEOF":                   bools,
	"KeepEmptyLines.AtEndOfFile":           bools,
	"KeepEmptyLines.AtStartOfBlock":        bools,
	"KeepEmptyLines.AtStartOfFile":         bools,
	"LambdaBodyIndentation":                {"Signature", "OuterScope"},
	"Language":                             {"Cpp"},      // this needs to be this specific value
	"LineEnding":                           {"DeriveLF"}, // this is locked to DeriveLF
	"MacroBlockBegin":                      {"''"},
	"MacroBlockEnd":                        {"''"},
	"MainIncludeChar":                      {"Any"}, // Locked to this value because messing with includes is bad
	"MaxEmptyLinesToKeep":                  {"0", "1", "2", "3", "4"},
	"PenaltyBreakAssignment":               {"2"}, // Penalties are left as is for now
	"PenaltyBreakBeforeFirstCallParameter": {"19"},
	"PenaltyBreakComment":                  {"300"},
	"PenaltyBreakFirstLessLess":            {"120"},
	"PenaltyBreakOpenParenthesis":          {"0"},
	"PenaltyBreakScopeResolution":          {"500"},
	"PenaltyBreakString":                   {"1000"},
	"PenaltyBreakTemplateDeclaration":      {"10"},
	"PenaltyExcessCharacter":               {"1000000"},
	"PenaltyIndentedWhitespace":            {"0"},
	"PenaltyReturnTypeOnItsOwnLine":        {"60"},
	"PointerAlignment":                     {"Right"}, // this is specifically called out in the confluence as Right
	"PPIndentWidth":                        {"-1"},    // -1 is IndentWidthfor preprocessor statements
	"QualifierAlignment":                   {"Leave"}, // locked to Leave, called out in docs that this is dangerous
	"ReferenceAlignment":                   {"Pointer", "Left", "Right", "Middle"},
	"ReflowComments":                       bools,
	"RemoveBracesLLVM":                     bools,
	"RemoveParentheses":                    {"Leave", "MultipleParentheses", "ReturnStatement"},
	"RemoveSemicolon":                      bools,
	"SeparateDefinitionBlocks":             {"Always", "Leave", "Never"},
	"SkipMacroDefinitionBody":              bools,
	"SortIncludes":                         {"Never"}, // locked to Never, changing this will break everything
	"SpaceAfterCStyleCast":                 bools,
	"SpaceAfterLogicalNot":                 bools,
	"SpaceAroundPointerQualifiers":         {"Default", "Before", "After", "Both"},
	"SpaceBeforeAssignmentOperators":       {"true"},
	"SpaceBeforeCaseColon":                 {"false"},
	"SpaceBeforeParens":                    {"Custom"},
	// locked into Custom to allow the SpaceBeforeParensOptions sub options to take effect
	"SpaceBeforeParensOptions.AfterControlStatements":       bools,
	"SpaceBeforeParensOptions.AfterForeachMacros":           bools,
	"SpaceBeforeParensOptions.AfterFunctionDefinitionName":  bools,
	"SpaceBeforeParensOptions.AfterFunctionDeclarationName": bools,
	"SpaceBeforeParensOptions.AfterIfMacros":                bools,
	"SpaceBeforeParensOptions.AfterOverloadedOperator":      bools,
	"SpaceBeforeParensOptions.AfterPlacementOperator":       bools,
	"SpaceBeforeParensOptions.AfterRequiresInClause":        bools,
	"SpaceBeforeParensOptions.AfterRequiresInExpression":    bools,
	"SpaceBeforeParensOptions.BeforeNonEmptyParentheses":    bools,
	"SpaceBeforeRangeBasedForLoopColon":                     bools,
	"SpaceBeforeSquareBrackets":                             bools,
	"SpaceInEmptyBlock":                                     bools,
	"SpacesInLineCommentPrefix.Minimum":                     {"1"},
	"SpacesInLineCommentPrefix.Maximum":                     {"1"},
	"SpacesInParens":                                        {"Custom"},
	// Needs to be custom so the SpacesInParensOptions sub options can work
	"SpacesInParensOptions.ExceptDoubleParentheses": bools,
	"SpacesInParensOptions.InConditionalStatements": bools,
	"SpacesInParensOptions.InCStyleCasts":           bools,
	"SpacesInParensOptions.InEmptyParentheses":      bools,
	"SpacesInParensOptions.Other":                   bools,
	"SpacesInSquareBrackets":                        bools,
	"TabWidth":                                      {"4"},
	"UseTab":                                        {"Never"},
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
			fmt.Sprintf("%s: %s\n", key, lines[key]),
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
	linesChangedTotal := math.MaxInt32
	var err error

	for j := 0; j < 2; j++ {
		fmt.Printf("Running iteration %d\n", j+1)

		format, linesChangedTotal, err = optimizeOptions(format, options, linesChangedTotal)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "optimizeOptions in iteration %d", j)
		}
	}

	// Let's go around the doublecheckafter bits
	fmt.Println("Running the doublechecks after")

	format, linesChangedTotal, err = optimizeOptions(format, doubleCheckAfter, linesChangedTotal)
	if err != nil {
		return nil, 0, errors.Wrap(err, "optimizeOptions in doubleCheck")
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
	if err != nil && !errors.Is(err, errNoLinesChanged) {
		return 0, errors.Wrap(err, "parseNumStat")
	}

	if errors.Is(err, errNoLinesChanged) {
		fmt.Println("No lines changed apparently...")
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
		return 0, errNoLinesChanged
	}

	// Split the input by lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return 0, errNoLinesChanged
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

func didLinesChange(in map[string]int) bool {
	lc := make([]int, len(in))
	i := 0
	for _, v := range in {
		lc[i] = v
		i++
	}

	previous := lc[0]
	for j := 0; j < len(lc)-1; j++ {
		if previous != lc[j+1] {
			return false
		}

		previous = lc[j+1]
	}

	return true
}

var doubleCheckAfter = map[string][]string{
	"AlignTrailingComments.OverEmptyLines": {"0", "1", "2", "3"},
	"ConstructorInitializerIndentWidth":    {"4", "2"},
	"ContinuationIndentWidth":              {"4", "2"},
	"IndentWidth":                          {"4", "2"},
	"MaxEmptyLinesToKeep":                  {"0", "1", "2", "3", "4"},
	"SpacesInLineCommentPrefix.Minimum":    {"0", "1", "2", "3"},
	"SpacesInLineCommentPrefix.Maximum":    {"0", "1", "2", "3"},
	"TabWidth":                             {"2", "4"},
}

func optimizeOptions(baseFormat ClangFormat, options map[string][]string, linesChangedTotal int) (ClangFormat, int,
	error) {
	// let's create a slice of option names
	optionNames := make([]string, len(options))
	i := 0
	for opt := range options {
		optionNames[i] = opt
		i++
	}

	// sort list alphabetically
	slices.Sort(optionNames)

	irrelevant := make([]string, 0)

	// for each option, let's check whether their individual values
	// would produce a diff that has a lower changed line count.
	for _, optionName := range optionNames {
		fmt.Printf(""+
			"==================%s\n"+
			"Checking option '%s'\n", strings.Repeat("=", len(optionName)), optionName)
		if len(options[optionName]) < 2 {
			fmt.Printf("Option '%s' is too short\n", optionName)
			continue
		}

		changes := make(map[string]int)

		for _, value := range options[optionName] {
			fmt.Printf("  Checking value\n"+
				"  %s: %s\n", optionName, value)
			if value == "" {
				panic(fmt.Sprintf("why %s", optionName))
			}
			baseFormat[optionName] = value

			linesChanged, err := runOption(baseFormat)
			if err != nil {
				return nil, 0, errors.Wrap(err, "runOption")
			}

			changes[value] = linesChanged
		}

		if didLinesChange(changes) {
			irrelevant = append(irrelevant, optionName)
		}

		minLinesChanged := math.MaxInt32
		winningValue := ""
		for value, linesChanged := range changes {
			if linesChanged < minLinesChanged {
				winningValue = value
				minLinesChanged = linesChanged
			}
		}

		fmt.Printf(""+
			"* Winning value was '%s' *\n"+
			"* with lines changed %d *\n",
			winningValue,
			minLinesChanged,
		)

		if linesChangedTotal > minLinesChanged {
			linesChangedTotal = minLinesChanged
		}

		baseFormat[optionName] = winningValue
	}

	fmt.Printf("These options did not have an effect on number of lines"+
		"changed whatever their value was: %v\n", irrelevant)

	return baseFormat, linesChangedTotal, nil
}
