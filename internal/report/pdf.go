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
	if err := r.GenerateTypstReport(&typstBuf); err != nil {
		return fmt.Errorf("failed to generate Typst content: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create Typst CLI compiler
	typstCLI := typst.CLI{}

	// Compile Typst to PDF
	options := &typst.CLIOptions{
		Format: typst.OutputFormatPDF,
	}

	if err := typstCLI.Compile(&typstBuf, outFile, options); err != nil {
		// Check if typst is not installed
		if strings.Contains(err.Error(), "executable file not found") {
			return fmt.Errorf("typst compiler not found in PATH. Please install Typst:\n  - Arch Linux: pacman -S typst\n  - macOS: brew install typst\n  - Cargo: cargo install typst-cli\n  - Or download from: https://github.com/typst/typst/releases")
		}
		return fmt.Errorf("PDF compilation failed: %w", err)
	}

	return nil
}
