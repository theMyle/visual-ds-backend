package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL_IPV4")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:5432/visualds?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Update category ordering
	categories := map[string]int{
		"introduction": 1,
		"big-o":        2,
		"array":        3,
		"linkedlist":   4,
		"stack":        5,
		"queue":        6,
	}

	for slug, order := range categories {
		_, err := db.Exec("UPDATE lesson_categories SET order_index = $1 WHERE slug = $2", order, slug)
		if err != nil {
			log.Printf("Failed to update order for %s: %v", slug, err)
		} else {
			fmt.Printf("Updated order for %s to %d\n", slug, order)
		}
	}

	// Update lesson ordering for introduction
	// 1. introduction-to-dsa
	// 2. what-is-an-algorithm
	// 3. what-is-a-ds
	lessonsIntro := map[string]int{
		"introduction-to-dsa":  1,
		"what-is-an-algorithm": 2,
		"what-is-a-ds":         3,
	}

	for slug, order := range lessonsIntro {
		_, err := db.Exec("UPDATE lessons SET order_index = $1 WHERE slug = $2", order, slug)
		if err != nil {
			log.Printf("Failed to update order for lesson %s: %v", slug, err)
		} else {
			fmt.Printf("Updated order for lesson %s to %d\n", slug, order)
		}
	}

	// Fetch some progress entries to see what's wrong
	rows, err := db.Query("SELECT user_id, lesson_category, lesson_id FROM user_lesson_progress LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("\n--- Progress Entries ---")
	for rows.Next() {
		var uid string
		var cat string
		var lid string
		if err := rows.Scan(&uid, &cat, &lid); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %s, Cat: %s, LessonID: %s\n", uid, cat, lid)
	}

	// Fetch categories to see lesson_count
	rows2, err := db.Query("SELECT lc.slug, COUNT(l.lesson_id) as lesson_count FROM lesson_categories lc LEFT JOIN lessons l ON lc.category_id = l.category_id GROUP BY lc.category_id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Println("\n--- Categories and Lesson Counts ---")
	for rows2.Next() {
		var slug string
		var count int
		if err := rows2.Scan(&slug, &count); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Category: %s, Lesson Count: %d\n", slug, count)
	}
}
