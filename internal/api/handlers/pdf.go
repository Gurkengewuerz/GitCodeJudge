package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	appConfig "github.com/gurkengewuerz/GitCodeJudge/internal/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

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
	tasks, err := judge.FindAllTasks(appCfg.TestPath)
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
	workshopMap := make(map[string][]judge.WorkshopTask)
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

			pdfURL := fmt.Sprintf("%s/pdf?task=%s/%s", appCfg.BaseURL, task.Workshop, task.Task)

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
	err = addPDFFooter(m, appCfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating footer")
	}

	return generateAndSendPDF(c, m)
}

func generateTaskPDF(c fiber.Ctx, appCfg *appConfig.Config, configPath string) error {
	parts := strings.Split(configPath, "/")
	if len(parts) != 2 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid task path")
	}

	workshopTask, err := judge.LoadWorkshopTask(appCfg.TestPath, parts[0], parts[1])
	if err != nil {
		log.WithError(err).Error("Error loading workshop task")
		return c.Status(fiber.StatusNotFound).SendString("Task not found")
	}

	// Check if problem is disabled or out of date range
	now := time.Now()
	if workshopTask.Config.Disabled ||
		(workshopTask.Config.StartDate != nil && now.Before(*workshopTask.Config.StartDate)) ||
		(workshopTask.Config.EndDate != nil && now.After(*workshopTask.Config.EndDate)) {
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

	// Add footer
	if err := addPDFFooter(m, appCfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating footer")
	}

	// Add content
	if err := addTaskContent(m, workshopTask); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error adding content")
	}

	return generateAndSendPDF(c, m)
}

func addPDFFooter(m core.Maroto, appCfg *appConfig.Config) error {
	return m.RegisterFooter(row.New(20).Add(
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
}

func addTaskContent(m core.Maroto, task *judge.WorkshopTask) error {
	// Add title
	m.AddRows(
		text.NewRow(10, task.Config.Name, props.Text{
			Top:   3,
			Style: fontstyle.Bold,
			Size:  16,
			Align: align.Center,
		}),
	)

	// Add solution file path information
	solutionPaths := []string{
		fmt.Sprintf("%s/%s/solution.<extension>", task.Workshop, task.Task),
	}

	m.AddRow(7, text.NewCol(12, "Solution File Path:", props.Text{
		Top:   2,
		Size:  10,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))

	for _, path := range solutionPaths {
		m.AddRow(5, text.NewCol(12, path, props.Text{
			Family: "Courier",
			Size:   9,
			Align:  align.Left,
			Color:  &props.Color{Red: 0, Green: 0, Blue: 200},
		}))
	}

	m.AddRow(7, text.NewCol(12, "Create a file in this path in your repository.", props.Text{
		Top:   1,
		Size:  9,
		Style: fontstyle.Italic,
		Align: align.Left,
	}))

	// Add description
	m.AddRow(10, text.NewCol(12, "Description:", props.Text{
		Top:   2,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))

	parts := strings.Split(task.Config.Description, "\n")
	for _, part := range parts {
		m.AddRow(0, text.NewCol(12, part, props.Text{
			Top:   0,
			Size:  10,
			Align: align.Left,
		}))
	}

	// Add date information
	if task.Config.StartDate != nil {
		m.AddRow(7, text.NewCol(12, "Available from: "+task.Config.StartDate.Format(time.RFC850), props.Text{
			Top:   5,
			Size:  9,
			Align: align.Left,
		}))
	}

	if task.Config.EndDate != nil {
		m.AddRow(7, text.NewCol(12, "Available until: "+task.Config.EndDate.Format(time.RFC850), props.Text{
			Top:   1,
			Size:  9,
			Align: align.Left,
		}))
	}

	// Example header
	m.AddRow(7, text.NewCol(12, "Examples", props.Text{
		Top:   2,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))

	for i, cases := range task.Config.Cases {
		// Test case header
		m.AddRow(7, text.NewCol(12, fmt.Sprintf("Example %d:", i+1), props.Text{
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
		inputLines := strings.Split(strings.TrimSpace(cases.Input), "\n")
		for _, s := range inputLines {
			m.AddRow(5, text.NewCol(12, s, props.Text{
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
		outputLines := strings.Split(judge.FormatExpectedString(cases.Expected), "\n")
		for _, s := range outputLines {
			m.AddRow(5, text.NewCol(12, s, props.Text{
				Family: "Courier",
				Size:   9,
				Align:  align.Left,
			}))
		}

		m.AddRow(0,
			line.NewCol(12, props.Line{}),
		)
	}

	return nil
}

func generateAndSendPDF(c fiber.Ctx, m core.Maroto) error {
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
