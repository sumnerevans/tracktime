package report

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Dadido3/go-typst"
)

// GeneratePDFReport generates a PDF report by compiling Typst to PDF
func (r *Report) GeneratePDFReport(outputPath string) error {
	// Generate Typst content to buffer
	var typstBuf bytes.Buffer
	r.GenerateTypstReport(&typstBuf)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create Typst CLI compiler with configured path
	typstCLI := typst.CLI{
		ExecutablePath: r.Config.TypstPath,
	}

	// Compile Typst to PDF
	options := &typst.CLIOptions{
		Format: typst.OutputFormatPDF,
	}

	if err := typstCLI.Compile(&typstBuf, outFile, options); err != nil {
		// Check if typst is not installed
		if strings.Contains(err.Error(), "executable file not found") {
			return fmt.Errorf("typst compiler not found in PATH")
		}
		return fmt.Errorf("PDF compilation failed: %w", err)
	}

	return nil
}
