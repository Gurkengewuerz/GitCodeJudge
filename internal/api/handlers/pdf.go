package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	appConfig "github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WorkshopTask struct {
	Workshop   string
	Task       string
	Config     models.TestCaseConfig
	ConfigPath string
}

func HandlePDF(appCfg *appConfig.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		configName := c.Query("task")
		if configName == "" {
			return generateWorkshopList(c, appCfg)
		}

		return generateTaskPDF(c, appCfg, configName)
	}
}

func generateWorkshopList(c fiber.Ctx, appCfg *appConfig.Config) error {
	tasks, err := findAllTasks(appCfg.TestPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error reading tasks")
	}

	// Create PDF configuration
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	// Create new PDF instance
	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	// Add title
	m.AddRows(
		text.NewRow(10, "Available Workshops and Tasks", props.Text{
			Top:   3,
			Style: fontstyle.Bold,
			Size:  16,
			Align: align.Center,
		}),
	)

	// Group tasks by workshop
	workshopMap := make(map[string][]WorkshopTask)
	for _, task := range tasks {
		workshopMap[task.Workshop] = append(workshopMap[task.Workshop], task)
	}

	// Add workshop sections
	for workshop, workshopTasks := range workshopMap {
		// Workshop header
		m.AddRow(10, text.NewCol(12, workshop, props.Text{
			Top:   2,
			Size:  14,
			Style: fontstyle.Bold,
			Align: align.Left,
		}))

		// Tasks in this workshop
		for _, task := range workshopTasks {
			if task.Config.Disabled {
				continue
			}

			now := time.Now()
			if (task.Config.StartDate != nil && now.Before(*task.Config.StartDate)) ||
				(task.Config.EndDate != nil && now.After(*task.Config.EndDate)) {
				continue
			}

			pdfURL := fmt.Sprintf("%s/pdf?task=%s/%s", c.BaseURL(), task.Workshop, task.Task)

			m.AddRow(7,
				text.NewCol(8, task.Config.Name, props.Text{
					Top:   1,
					Size:  10,
					Align: align.Left,
				}),
				text.NewCol(4, pdfURL, props.Text{
					Top:   1,
					Size:  8,
					Style: fontstyle.Italic,
					Align: align.Right,
					Color: &props.Color{Red: 0, Green: 0, Blue: 200},
				}),
			)

			if task.Config.StartDate != nil && task.Config.EndDate != nil {
				m.AddRow(5,
					text.NewCol(12, fmt.Sprintf("Available: %s - %s",
						task.Config.StartDate.Format(time.RFC850),
						task.Config.EndDate.Format(time.RFC850)), props.Text{
						Top:   1,
						Size:  8,
						Style: fontstyle.Italic,
						Align: align.Left,
						Color: &props.Color{Red: 100, Green: 100, Blue: 100},
					}),
				)
			}
		}
	}

	// Add footer
	err = m.RegisterFooter(row.New(20).Add(
		col.New(12).Add(
			text.New(appCfg.PDFFooterCopyright, props.Text{
				Top:   13,
				Style: fontstyle.Italic,
				Size:  8,
				Align: align.Left,
			}),
			text.New(appCfg.PDFFooterGeneratedWith, props.Text{
				Top:   16,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Left,
			}),
		),
	))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating footer")
	}

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating PDF")
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "workshops-*.pdf")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating temporary file")
	}
	defer os.Remove(tmpFile.Name())

	// Save PDF to temporary file
	if err := document.Save(tmpFile.Name()); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error saving PDF")
	}

	// Set content type and send file
	c.Set("Content-Type", "application/pdf")
	return c.SendFile(tmpFile.Name(), fiber.SendFile{
		Compress: true,
	})
}

func findAllTasks(testPath string) ([]WorkshopTask, error) {
	var tasks []WorkshopTask

	err := filepath.WalkDir(testPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "config.yaml" {
			// Read and parse config file
			yamlData, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip this file if we can't read it
			}

			var config models.TestCaseConfig
			if err := yaml.Unmarshal(yamlData, &config); err != nil {
				return nil // Skip this file if we can't parse it
			}

			// Get relative path components
			relPath, err := filepath.Rel(testPath, filepath.Dir(path))
			if err != nil {
				return nil // Skip if we can't get relative path
			}

			pathParts := strings.Split(relPath, string(os.PathSeparator))
			if len(pathParts) != 2 {
				return nil // Skip if path structure is not workshop/task
			}

			tasks = append(tasks, WorkshopTask{
				Workshop:   pathParts[0],
				Task:       pathParts[1],
				Config:     config,
				ConfigPath: path,
			})
		}

		return nil
	})

	return tasks, err
}

func generateTaskPDF(c fiber.Ctx, appCfg *appConfig.Config, configName string) error {
	// Ensure configName is safe to use in file path
	configName = filepath.Clean(configName)
	if strings.Contains(configName, "..") {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid config name")
	}

	// Read and parse YAML file
	configPath := filepath.Join(appCfg.TestPath, configName, "config.yaml")
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Config file not found")
	}

	var testConfig models.TestCaseConfig
	if err := yaml.Unmarshal(yamlData, &testConfig); err != nil {
		log.WithError(err).Error("Error parsing config file")
		return c.Status(fiber.StatusInternalServerError).SendString("Error parsing config file")
	}

	// Check if problem is disabled or out of date range
	now := time.Now()
	if testConfig.Disabled || (testConfig.StartDate != nil && now.Before(*testConfig.StartDate)) || (testConfig.EndDate != nil && now.After(*testConfig.EndDate)) {
		return c.Status(fiber.StatusNotFound).SendString("Problem not available")
	}

	// Create PDF configuration
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	// Create new PDF instance
	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err = m.RegisterFooter(row.New(20).Add(
		col.New(12).Add(
			text.New(appCfg.PDFFooterCopyright, props.Text{
				Top:   13,
				Style: fontstyle.Italic,
				Size:  8,
				Align: align.Left,
			}),
			text.New(appCfg.PDFFooterGeneratedWith, props.Text{
				Top:   16,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Left,
			}),
		),
	))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("failed to create footer")
	}

	// Add title
	m.AddRows(
		text.NewRow(10, testConfig.Name, props.Text{
			Top:   3,
			Style: fontstyle.Bold,
			Size:  16,
			Align: align.Center,
		}),
	)

	// Add description
	parts := strings.Split(testConfig.Description, "\n")
	for _, part := range parts {
		m.AddRow(0, text.NewCol(12, part, props.Text{
			Top:   0,
			Size:  10,
			Align: align.Left,
		}))
	}

	// Add date information
	m.AddRow(7, text.NewCol(12, "Available from: "+testConfig.StartDate.Format(time.RFC850), props.Text{
		Top:   5,
		Size:  9,
		Align: align.Left,
	}),
	)

	m.AddRow(7, text.NewCol(12, "Available until: "+testConfig.EndDate.Format(time.RFC850), props.Text{
		Top:   1,
		Size:  9,
		Align: align.Left,
	}))

	// Add example section if test cases exist
	if len(testConfig.Cases) > 0 {
		// Example header
		m.AddRow(7, text.NewCol(12, "Example:", props.Text{
			Top:   2,
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Left,
		}))

		// Input section
		m.AddRow(7, text.NewCol(12, "Input:", props.Text{
			Top:   1,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Left,
		}))

		// Format input with monospace font
		inputLines := strings.Split(strings.TrimSpace(testConfig.Cases[0].Input), "\n")
		for _, line := range inputLines {
			m.AddRow(5, text.NewCol(12, "    "+line, props.Text{
				Family: "Courier",
				Size:   9,
				Align:  align.Left,
			}))
		}

		// Expected output section
		m.AddRow(7, text.NewCol(12, "Expected Output:", props.Text{
			Top:   1,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Left,
		}))

		// Format expected output with monospace font
		outputLines := strings.Split(judge.FormatExpectedString(testConfig.Cases[0].Expected), "\n")
		for _, line := range outputLines {
			m.AddRow(5, text.NewCol(12, "    "+line, props.Text{
				Family: "Courier",
				Size:   9,
				Align:  align.Left,
			}))
		}
	}

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		log.WithError(err).Error("Error generating PDF")
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating PDF")
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "problem-*.pdf")
	if err != nil {
		log.WithError(err).Error("Error creating temporary file")
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating temporary file")
	}
	defer os.Remove(tmpFile.Name())

	// Save PDF to temporary file
	if err := document.Save(tmpFile.Name()); err != nil {
		log.WithError(err).Error("Error saving PDF")
		return c.Status(fiber.StatusInternalServerError).SendString("Error saving PDF")
	}

	// Set content type and send file
	c.Set("Content-Type", "application/pdf")
	return c.SendFile(tmpFile.Name(), fiber.SendFile{
		Compress: true,
	})
}
