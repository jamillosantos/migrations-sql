package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	colorHighlight = color.New(color.Bold, color.FgHiWhite).Sprint
	colorError     = color.New(color.FgRed).Sprint
)

var (
	destination = "."
	extension   = "sql"
	withUndo    = false
	withDown    = false
	format      = "20060102150405"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [--destination=<destination>] [--extension=sql] [description]",
	Short: "Creates a new migration files.",
	Long: `
Examples:
To create the a migration:
$ migrations create --destination=./migrations Create customers table
If the description is not provided, the this command will ask for it.
$ migrations create --destination=./migrations
`,
	Example: `migrations create --destination=./migrations Create table transactions`,
	Run: func(cmd *cobra.Command, args []string) {
		var description string
		if len(args) == 0 {
			err := survey.AskOne(&survey.Input{
				Message: "Type the description of the migration:",
			}, &description)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		} else {
			description = strings.Join(args, " ")
		}

		_, err := os.Stat(destination)
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(destination, 0750)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		var id string
		switch format {
		case "unix":
			id = strconv.FormatInt(time.Now().UTC().Unix(), 10)
		default:
			id = time.Now().UTC().Format(format)
		}
		var suffix string
		if withDown {
			suffix = "up"
		} else if withUndo {
			suffix = "do"
		}
		migrationFilePath := []string{path.Join(destination, migrationFileName(id, description, suffix, extension))}
		if withDown {
			suffix = "down"
		} else if withUndo {
			suffix = "undo"
		}
		if suffix != "" {
			migrationFilePath = append(migrationFilePath, path.Join(destination, migrationFileName(id, description, suffix, extension)))
		}

		for _, m := range migrationFilePath {
			err := createMigrationFile(m)
			if err != nil {
				os.Exit(1)
			}
			fmt.Printf("migration %s was created", colorHighlight(m))
			fmt.Println()
		}
	},
}

func createMigrationFile(migrationFilePath string) error {
	f, err := os.Create(migrationFilePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed creating the migration file (%s): %s", migrationFilePath, colorError(err.Error()))
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	return nil
}

func migrationFileName(id string, description string, suffix string, e string) string {
	if suffix != "" {
		suffix = "." + suffix
	}
	return fmt.Sprintf(
		"%s_%s%s.%s",
		id,
		strings.Map(func(r rune) rune {
			if r == 32 /* space */ {
				return '_'
			} else if unicode.IsPrint(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, description),
		suffix,
		e,
	)
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&destination, "destination", "d", destination, "Folder where the migrations file will be created")
	createCmd.Flags().StringVarP(&format, "format", "f", format, "Format of the migration ID (default, unix, any time.Time.Format supported)")
	createCmd.Flags().StringVarP(&extension, "extension", "e", extension, "Extension of the migration that will be created")
	createCmd.Flags().BoolVar(&withUndo, "undo", withUndo, "Enable undo file")
	createCmd.Flags().BoolVar(&withDown, "down", withDown, "Enable down file")
	createCmd.Flags().BoolVarP(&withUndo, "interactive", "i", withUndo, "Enable interactive mode")
}
