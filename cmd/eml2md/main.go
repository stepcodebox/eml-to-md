package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/stepcodebox/eml-to-md/internal/converter"
)

type config struct {
	outputDir string
	stdout    bool
	recursive bool
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cfg, inputs, err := parseArgs(args)
	if err != nil {
		return err
	}

	if len(inputs) == 0 {
		return errors.New("no input files; pass .eml files, directories, or - for stdin")
	}

	expanded, err := expandInputs(inputs, cfg.recursive)
	if err != nil {
		return err
	}

	if len(expanded) == 0 {
		return errors.New("no .eml files found")
	}

	var failures int
	for _, input := range expanded {
		if err := convertOne(input, cfg, stdin, stdout); err != nil {
			failures++
			fmt.Fprintf(stderr, "error: %s: %v\n", input, err)
			continue
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d conversion(s) failed", failures)
	}
	return nil
}

func parseArgs(args []string) (config, []string, error) {
	var cfg config
	fs := flag.NewFlagSet("eml2md", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.outputDir, "o", "", "output directory")
	fs.StringVar(&cfg.outputDir, "output-dir", "", "output directory")
	fs.BoolVar(&cfg.stdout, "stdout", false, "write Markdown to stdout")
	fs.BoolVar(&cfg.recursive, "r", false, "recursively scan directories")
	fs.BoolVar(&cfg.recursive, "recursive", false, "recursively scan directories")

	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}
	return cfg, fs.Args(), nil
}

func expandInputs(inputs []string, recursive bool) ([]string, error) {
	var out []string
	for _, input := range inputs {
		if input == "-" {
			out = append(out, input)
			continue
		}

		info, err := os.Stat(input)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			out = append(out, input)
			continue
		}

		if !recursive {
			entries, err := os.ReadDir(input)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".eml") {
					out = append(out, filepath.Join(input, entry.Name()))
				}
			}
			continue
		}

		err = filepath.WalkDir(input, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".eml") {
				out = append(out, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func convertOne(input string, cfg config, stdin io.Reader, stdout io.Writer) error {
	var data []byte
	var err error
	if input == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(input)
	}
	if err != nil {
		return err
	}

	markdown, err := converter.Convert(bytes.NewReader(data))
	if err != nil {
		return err
	}

	if cfg.stdout || input == "-" {
		_, err = stdout.Write([]byte(markdown))
		return err
	}

	outputPath := outputPath(input, cfg.outputDir)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(markdown), 0o644)
}

func outputPath(input string, outputDir string) string {
	name := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input)) + ".md"
	if outputDir == "" {
		return filepath.Join(filepath.Dir(input), name)
	}
	return filepath.Join(outputDir, name)
}
