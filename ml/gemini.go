package ml

import (
	"NUMParser/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/genai"
)

var (
	titleRegex  = regexp.MustCompile(`(.+?).\((\d\d\d\d)\);?`)
	GoogleAiKey = ""
	err         error
)

func Init() {

	log.Println("Init ml")

	GoogleAiKey, err = config.ReadConfigParser("AigKey")

	if err != nil || GoogleAiKey == "" {
		dir := filepath.Dir(os.Args[0])
		buf, err := os.ReadFile(filepath.Join(dir, "aig.key"))
		if err != nil || strings.TrimSpace(string(buf)) == "" {
			log.Println("Fatal error read google ai key:", err)
			os.Exit(1)
		}
		GoogleAiKey = strings.TrimSpace(string(buf))
	}

	LoadConfig()

	if CollsConfig == nil || len(CollsConfig.Collections) < 50 {
		if CollsConfig == nil {
			CollsConfig = &CollectionsConfig{
				Collections: nil,
				LastUpdated: time.Now(),
			}
		}
		log.Println("Generate started collections")
		for len(CollsConfig.Collections) < 50 {
			colls, err := GenCollection(50 - len(CollsConfig.Collections))
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 5)
				continue
			}
			for _, coll := range colls {
				fmt.Println(coll.Title)
				fmt.Println(coll.Overview)
				fmt.Println(coll.Prompt)
				fmt.Println()
			}
			CollsConfig.Collections = colls
			// при создании файлы нужно собрать файл коллекций
			CollsConfig.LastUpdated = time.Now().Add(-time.Hour * 200)
		}
		SaveConfig()
	}
}

func GetCollectionMovies(collection *Collection) ([]*MovieInfo, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  GoogleAiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	fullPrompt := `Ты кинокритик. Дай подборку фильмов. Ответ должен быть только в виде списка названий, разделенных ТОЛЬКО точкой с запятой. 
Правила:
1. Используй только оригинальные названия
2. Указывай год выпуска
3. Используй только реально существующие фильмы
4. Формат: Title1 (Year1); Title2 (Year2); Title3 (Year3); ...
5. Не добавляй номера, точки, кавычки или другие символы
6. Не более 30 фильмов
7. Короткометражные фильмы и сериалы не включать в список
8. Строго запрещено повторять названия фильмов в ответе
9. Кино произведенное в СССР или России должны быть названия на русском
` + collection.Prompt

	contents := []*genai.Content{
		genai.NewContentFromText(fullPrompt, genai.RoleUser),
	}

	//resp, err := client.Models.GenerateContent(ctx, "gemma-3-27b-it", contents, nil)
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, nil)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("Gemini returned empty response")
	}

	var generatedTextBuilder strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				generatedTextBuilder.WriteString(string(part.Text))
			}
		}
	}

	fmt.Println("Resp:", generatedTextBuilder.String())
	return parseTitles(generatedTextBuilder.String()), nil
}

func GenCollection(count int) ([]*Collection, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  GoogleAiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	collectionsJSON := ""
	if CollsConfig != nil && len(CollsConfig.Collections) > 0 {
		// Преобразуем существующие коллекции в JSON для отправки в промпт
		buf, err := json.Marshal(CollsConfig.Collections)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal existing collections: %w", err)
		}
		collectionsJSON = `, темы которой нет в этом списке: ` + string(buf) + "."
	} else {
		collectionsJSON = `, пример подборок: [{"Наши друзья", "Семейные фильмы про животных", "Фильмы про кошек, собак и других животных друзей людей, начиная с 1990 года и жанрами комедия, семейный фильм", ""},
		{"Космическая одиссея", "Классика фантастики о покорении космоса", "Научно-фантастические фильмы про освоение космоса до 2000 года", ""},
		{"Вселенная Нолана", "Хронология фильмов культового режиссера", "Фильмы Кристофера Нолана в хронологическом порядке", ""},
		{"Ностальгические 90-е", "Знаковые подростковые комедии эпохи", "Популярные подростковые комедии 1995-1999 годов", ""}]`
	}

	prompt := `
Ответь только JSON-кодом. Сгенерируй новую, уникальную коллекцию фильмов` + collectionsJSON + `.
Используй формат массива JSON-объектов, где каждый объект имеет ключи: 	Title (string), Overview (string), Prompt (string). 
Не добавляй никаких других ключей или текста, кроме самого JSON.
В списке должно содержаться ` + strconv.Itoa(count) + ` коллекций.
Старайся в подборку включать современные фильмы, если позволяет подборка.
`

	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, nil)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("Gemini returned empty response")
	}

	var generatedTextBuilder strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				generatedTextBuilder.WriteString(string(part.Text))
			}
		}
	}

	generatedText := generatedTextBuilder.String()

	// Gemini может иногда добавлять обратные кавычки `
	generatedText = strings.TrimPrefix(generatedText, "```json\n")
	generatedText = strings.TrimSuffix(generatedText, "```")
	generatedText = strings.TrimSpace(generatedText)

	var collections []*Collection
	err = json.Unmarshal([]byte(generatedText), &collections)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON from Gemini response: %w", err)
	}

	return collections, nil
}

func parseTitles(response string) []*MovieInfo {
	matches := titleRegex.FindAllStringSubmatch(response, -1)
	var movies []*MovieInfo
	for _, match := range matches {
		if len(match) > 2 {
			title := cleanTitle(match[1])
			year := match[2]
			movies = append(movies, &MovieInfo{Title: title, Year: year})
		}
	}
	return movies
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	return strings.Trim(title, `"'`)
}
