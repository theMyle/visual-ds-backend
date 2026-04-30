package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"visualds/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL_IPV4")
	if dbURL == "" {
		// Fallback to local if not set
		dbURL = "postgresql://postgres:postgres@localhost:5432/visualds?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := database.New(db)
	ctx := context.Background()

	// Adjust this path to point to your frontend content directory
	baseDir := filepath.Join("..", "frontend", "src", "content", "lessons")
	
	categories, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Fatalf("failed to read base directory: %v", err)
	}

	titleRegex := regexp.MustCompile(`(?s)---\s*title:\s*(.*?)\s*---`)

	for i, catDir := range categories {
		if !catDir.IsDir() {
			continue
		}

		catSlug := catDir.Name()
		catPath := filepath.Join(baseDir, catSlug)
		
		fmt.Printf("Processing category: %s\n", catSlug)

		// Create or get category
		category, err := queries.GetLessonCategoryBySlug(ctx, catSlug)
		if err != nil {
			if err == sql.ErrNoRows {
				category, err = queries.CreateLessonCategory(ctx, database.CreateLessonCategoryParams{
					Slug:        catSlug,
					Title:       strings.Title(strings.ReplaceAll(catSlug, "-", " ")),
					Description: sql.NullString{Valid: false},
					OrderIndex:  int32(i),
				})
				if err != nil {
					log.Printf("failed to create category %s: %v", catSlug, err)
					continue
				}
			} else {
				log.Printf("failed to check category %s: %v", catSlug, err)
				continue
			}
		}

		lessons, err := ioutil.ReadDir(catPath)
		if err != nil {
			log.Printf("failed to read category directory %s: %v", catSlug, err)
			continue
		}

		for j, lessonFile := range lessons {
			if lessonFile.IsDir() || !strings.HasSuffix(lessonFile.Name(), ".md") {
				continue
			}

			lessonSlug := strings.TrimSuffix(lessonFile.Name(), ".md")
			lessonPath := filepath.Join(catPath, lessonFile.Name())

			contentBytes, err := ioutil.ReadFile(lessonPath)
			if err != nil {
				log.Printf("failed to read lesson file %s: %v", lessonPath, err)
				continue
			}

			content := string(contentBytes)
			title := lessonSlug

			// Extract title from frontmatter
			match := titleRegex.FindStringSubmatch(content)
			if len(match) > 1 {
				title = strings.TrimSpace(match[1])
				// Remove frontmatter from content
				content = titleRegex.ReplaceAllString(content, "")
				content = strings.TrimSpace(content)
			}

			// Create or update lesson
			existingLesson, err := queries.GetLessonBySlug(ctx, database.GetLessonBySlugParams{
				Slug:   catSlug,
				Slug_2: lessonSlug,
			})

			if err != nil {
				if err == sql.ErrNoRows {
					_, err = queries.CreateLesson(ctx, database.CreateLessonParams{
						CategoryID: category.CategoryID,
						Slug:       lessonSlug,
						Title:      title,
						Content:    content,
						OrderIndex: int32(j),
					})
					if err != nil {
						log.Printf("failed to create lesson %s: %v", lessonSlug, err)
					} else {
						fmt.Printf("  Created lesson: %s\n", lessonSlug)
					}
				} else {
					log.Printf("failed to check lesson %s: %v", lessonSlug, err)
				}
			} else {
				_, err = queries.UpdateLesson(ctx, database.UpdateLessonParams{
					LessonID:   existingLesson.LessonID,
					Title:      title,
					Content:    content,
					OrderIndex: int32(j),
				})
				if err != nil {
					log.Printf("failed to update lesson %s: %v", lessonSlug, err)
				} else {
					fmt.Printf("  Updated lesson: %s\n", lessonSlug)
				}
			}
		}
	}

	fmt.Println("Seeding completed!")
}
